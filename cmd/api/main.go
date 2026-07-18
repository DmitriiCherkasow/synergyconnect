package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("🚀 SynergyConnect starting...")

	// Простой health-check эндпоинт
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "service": "synergyconnect"}`))
	})

	log.Println("✅ Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
