package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/store"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	// Driver SQLite3 harus di-import
	_ "github.com/mattn/go-sqlite3"
)

// WAQRString menyimpan QR string terbaru untuk diambil oleh web server
var WAQRString string

// InitWhatsApp menginisialisasi client WhatsApp dengan whatsmeow
func InitWhatsApp(eventHandler whatsmeow.EventHandler) (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Pastikan folder data ada
	os.MkdirAll("data", os.ModePerm)
	// Kita akan menggunakan sqlite3 lokal untuk menyimpan session (cookie login WA)
	container, err := sqlstore.New(context.Background(), "sqlite3", "file:data/wa_session.db?_foreign_keys=on", dbLog)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungkan ke database sesi: %w", err)
	}

	// Ambil device pertama jika sudah pernah login, jika belum maka nil
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil perangkat: %w", err)
	}

	// Identifikasi secara persis sebagai perangkat Desktop / Chrome
	store.DeviceProps.PlatformType = store.DeviceProps_CHROME.Enum()
	store.DeviceProps.Os = proto.String("Mac OS")
	store.DeviceProps.RequireFullSync = proto.Bool(false)

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Daftarkan event handler untuk memproses pesan masuk dll
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		// Jika belum login, jalankan background routine untuk qr
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			return nil, fmt.Errorf("gagal connect: %w", err)
		}
		
		go func() {
			timer := time.NewTimer(3 * time.Minute)
			defer timer.Stop()

			for {
				select {
				case evt, ok := <-qrChan:
					if !ok {
						return
					}
					if evt.Event == "code" {
						WAQRString = evt.Code
						fmt.Println("\nQR Code baru diterima, silakan buka http://localhost:3000 untuk scan!")
					} else if evt.Event == "success" {
						fmt.Println("Login WhatsApp berhasil!")
						return
					} else {
						fmt.Println("Login event:", evt.Event)
					}
				case <-timer.C:
					fmt.Println("Waktu scan QR (3 menit) habis. Memutuskan koneksi...")
					client.Disconnect()
					WAQRString = ""
					return
				}
			}
		}()
	} else {
		// Jika sudah login, langsung connect saja
		err = client.Connect()
		if err != nil {
			return nil, fmt.Errorf("gagal connect: %w", err)
		}
		fmt.Println("Berhasil terhubung ke WhatsApp (menggunakan sesi yang ada)!")
	}

	return client, nil
}
