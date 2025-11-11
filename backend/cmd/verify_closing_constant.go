package main

import (
	"fmt"
	"app-sistem-akuntansi/models"
)

func main() {
	fmt.Println("=== Verifying Closing Constant ===")
	fmt.Printf("JournalRefClosing value: %q\n", models.JournalRefClosing)
	fmt.Println()
	
	if models.JournalRefClosing == "CLOSING" {
		fmt.Println("✅ SUKSES: Konstanta sudah benar!")
		fmt.Println("✅ Service akan menggunakan reference_type = 'CLOSING'")
		fmt.Println("✅ Ini akan match dengan data di database")
	} else {
		fmt.Printf("❌ ERROR: Konstanta masih salah: %q\n", models.JournalRefClosing)
		fmt.Println("❌ Seharusnya: 'CLOSING'")
	}
	
	fmt.Println("\n=== Expected Query ===")
	fmt.Printf("Query akan menjadi: WHERE reference_type = %q\n", models.JournalRefClosing)
}
