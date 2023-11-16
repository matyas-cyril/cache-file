package cacheFile

import (
	"crypto/aes"
	"crypto/cipher"
	crypto_rand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
)

// genAESKey: Générer une clef de 32 bytes
// peu importe la valeur d'entrée
func genAESKey(key []byte) []byte {
	k := sha256.Sum256(key)
	return k[:]
}

// randKey: Générer une clef aléatoire dont le nombre de carac
// est défini par length
func randKey(iteration uint) []byte {

	str := ""
	for i := 0; len(str) < int(iteration); i++ {
		nbr := make([]byte, 8)
		binary.LittleEndian.PutUint64(nbr, rand.Uint64())
		sum := sha256.Sum256(nbr)
		str = fmt.Sprintf("%s%s", str, base64.StdEncoding.EncodeToString(sum[:]))
	}

	tab := []byte(str)
	rand.Shuffle(len(tab), func(i, j int) {
		tab[i], tab[j] = tab[j], tab[i]
	})

	return tab[:32]
}

// encipher: chiffrer des données en fonction d'une clef
func encipher(data, key []byte) ([]byte, error) {

	key = genAESKey(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Génération d'un nbr arbitraitre
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(crypto_rand.Reader, nonce); err != nil {
		return nil, err
	}

	cipherdata := gcm.Seal(nonce, nonce, data, nil)

	return cipherdata, nil
}

// decrypt: déchiffrer des données en fonction d'une clef
func decipher(cipherdata, key []byte) ([]byte, error) {

	key = genAESKey(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherdata) < nonceSize {
		return nil, fmt.Errorf("data %d capacity < noncesize %d", len(cipherdata), nonceSize)
	}

	nonce, cipherdata := cipherdata[:nonceSize], cipherdata[nonceSize:]
	data, err := gcm.Open(nil, nonce, cipherdata, nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}
