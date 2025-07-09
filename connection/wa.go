// connection/wa.go
package connection

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var (
	client    *whatsmeow.Client
	container *sqlstore.Container
	clientMux sync.RWMutex
	waCtx     = context.Background()
)

func GetWAClient() *whatsmeow.Client {
	clientMux.RLock()
	defer clientMux.RUnlock()
	return client
}

func InitWAClient() error {
	clientMux.Lock()
	defer clientMux.Unlock()
	store.DeviceProps.Os = proto.String("AL WhatsApp")

	if client != nil {
		return nil
	}

	// Setup database dengan log level yang lebih rendah
	dbLog := waLog.Stdout("Database", "ERROR", true)
	var err error
	container, err = sqlstore.New(waCtx, "sqlite3", "file:wa_agent.db?_foreign_keys=on", dbLog)
	if err != nil {
		return err
	}

	// Get device store
	deviceStore, err := container.GetFirstDevice(waCtx)
	if err != nil {
		return err
	}

	// Create client dengan log level yang lebih rendah
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client = whatsmeow.NewClient(deviceStore, clientLog)

	if client.Store.ID != nil {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func ConnectWA() (<-chan whatsmeow.QRChannelItem, error) {
	clientMux.RLock()
	defer clientMux.RUnlock()

	if client == nil {
		return nil, nil
	}

	if client.IsConnected() {
		return nil, nil
	}

	// Get QR channel
	qrChan, err := client.GetQRChannel(context.Background())
	if err != nil {
		return nil, err
	}

	// Connect in background
	go client.Connect()

	return qrChan, nil
}

func DisconnectWA() error {
	clientMux.Lock()
	defer clientMux.Unlock()

	if client != nil {
		// Logout untuk membersihkan session di server WhatsApp
		if client.IsConnected() {
			err := client.Logout(waCtx)
			if err != nil {
				// Jika logout gagal, tetap lanjut disconnect
				client.Disconnect()
			}
		} else {
			client.Disconnect()
		}

		// Tunggu cleanup
		time.Sleep(2 * time.Second)

		// Clear device dari database untuk memastikan cleanup
		if client.Store != nil {
			err := client.Store.Delete(waCtx)
			if err != nil {
				// Log error tapi tetap lanjut
				// log.Printf("Failed to delete device from store: %v", err)
			}
		}

		client = nil
	}

	// Optional: Close container connection
	if container != nil {
		container.Close()
		container = nil
	}

	return nil
}

func ForceCleanup() error {
	clientMux.Lock()
	defer clientMux.Unlock()

	// Force disconnect jika masih ada client
	if client != nil {
		if client.IsConnected() {
			client.Disconnect()
		}
		// Delete device store
		if client.Store != nil {
			client.Store.Delete(waCtx)
		}
		client = nil
	}

	// Close container
	if container != nil {
		container.Close()
		container = nil
	}

	// Hapus database file
	dbPath := "wa_agent.db"
	if _, err := os.Stat(dbPath); err == nil {
		return os.Remove(dbPath)
	}

	return nil
}

func IsConnected() bool {
	clientMux.RLock()
	defer clientMux.RUnlock()

	if client == nil {
		return false
	}
	return client.IsConnected()
}

func GetUserID() string {
	clientMux.RLock()
	defer clientMux.RUnlock()

	if client == nil || client.Store.ID == nil {
		return ""
	}
	return client.Store.ID.User
}

// Fungsi untuk reset complete - gunakan ketika ada masalah
func ResetWAClient() error {
	clientMux.Lock()
	defer clientMux.Unlock()

	// Force disconnect
	if client != nil {
		if client.IsConnected() {
			client.Disconnect()
		}
		// Delete device store
		if client.Store != nil {
			client.Store.Delete(waCtx)
		}
		time.Sleep(1 * time.Second)
		client = nil
	}

	// Close container
	if container != nil {
		container.Close()
		container = nil
	}

	// Hapus database dan buat ulang
	dbPath := "wa_agent.db"
	os.Remove(dbPath)

	// Re-initialize
	dbLog := waLog.Stdout("Database", "ERROR", true)
	var err error
	container, err = sqlstore.New(waCtx, "sqlite3", "file:wa_agent.db?_foreign_keys=on", dbLog)
	if err != nil {
		return err
	}

	deviceStore, err := container.GetFirstDevice(waCtx)
	if err != nil {
		return err
	}

	clientLog := waLog.Stdout("Client", "ERROR", true)
	client = whatsmeow.NewClient(deviceStore, clientLog)

	return nil
}


// SendTextMessage mengirim pesan teks ke nomor WhatsApp
func SendTextMessage(to, message string) error {
	clientMux.RLock()
	defer clientMux.RUnlock()

	if client == nil {
		return fmt.Errorf("client not initialized")
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	// Format nomor WhatsApp
	jid, err := types.ParseJID(to)
	if err != nil {
		// Coba format nomor jika parsing gagal
		// Hapus karakter non-digit
		cleanNumber := ""
		for _, char := range to {
			if char >= '0' && char <= '9' {
				cleanNumber += string(char)
			}
		}
		
		// Tambahkan @s.whatsapp.net jika belum ada
		if !strings.Contains(cleanNumber, "@") {
			// Jika nomor dimulai dengan 0, ganti dengan 62 (Indonesia)
			if strings.HasPrefix(cleanNumber, "0") {
				cleanNumber = "62" + cleanNumber[1:]
			}
			cleanNumber += "@s.whatsapp.net"
		}
		
		jid, err = types.ParseJID(cleanNumber)
		if err != nil {
			return fmt.Errorf("invalid phone number format: %v", err)
		}
	}

	// Buat pesan
	msg := &waProto.Message{
		Conversation: proto.String(message),
	}

	// Kirim pesan
	_, err = client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}

// SendMessageWithRetry mengirim pesan dengan retry mechanism
func SendMessageWithRetry(to, message string, maxRetries int) error {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		err := SendTextMessage(to, message)
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Jika client tidak connected, jangan retry
		if strings.Contains(err.Error(), "not connected") {
			break
		}
		
		// Wait sebelum retry
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	
	return lastErr
}

// CheckNumber mengecek apakah nomor WhatsApp valid/terdaftar
func CheckNumber(phoneNumber string) (bool, error) {
	clientMux.RLock()
	defer clientMux.RUnlock()

	if client == nil {
		return false, fmt.Errorf("client not initialized")
	}

	if !client.IsConnected() {
		return false, fmt.Errorf("client not connected")
	}

	// Format nomor
	cleanNumber := ""
	for _, char := range phoneNumber {
		if char >= '0' && char <= '9' {
			cleanNumber += string(char)
		}
	}

	if strings.HasPrefix(cleanNumber, "0") {
		cleanNumber = "62" + cleanNumber[1:]
	}

	if !strings.HasPrefix(cleanNumber, "+") {
		cleanNumber = "+" + cleanNumber
	}

	if !strings.Contains(cleanNumber, "@") {
		cleanNumber += "@s.whatsapp.net"
	}

	jid, err := types.ParseJID(cleanNumber)
	if err != nil {
		return false, err
	}

	// Check if number is registered on WhatsApp
	resp, err := client.IsOnWhatsApp([]string{jid.User})
	if err != nil {
		return false, err
	}

	if len(resp) > 0 {
		return resp[0].IsIn, nil
	}

	return false, nil
}