package cacheFile_test

import (
	"crypto/sha512"
	"fmt"
	"math/rand"
	"testing"

	cacheFile "github.com/matyas-cyril/cache-file"
)

// Ici un exemple de génération d'un map[string][]byte et du hash
func createData(usr, dom, srv, pwd string, auth bool) ([64]byte, map[string][]byte) {

	// Exemple de données
	data := map[string][]byte{
		"usr": []byte(usr),
		"dom": []byte(dom),
		"srv": []byte(srv),
	}

	if auth {
		data["aut"] = []byte{1}
	} else {
		data["aut"] = []byte{0}
	}

	// Exemple de génération de la clef obligatoire
	hash := sha512.Sum512([]byte(fmt.Sprintf("%s%s%s%s", data["usr"], data["dom"], data["srv"], pwd)))

	return hash, data
}

// TestCreateCache: créer 100 fichiers avec une durée max de validité de 60 sec
// go test -timeout 3s -run ^TestCreateCache$
func TestCreateCache(t *testing.T) {

	c, err := cacheFile.New("/tmp/test")
	if err != nil {
		t.Fatal(err.Error())
	}

	nbrFile := 100

	for i := 0; i < nbrFile; i++ {
		hash, data := createData(fmt.Sprintf("user%d", i), "test.fr", "imap", "mot_de_passe", true)

		// Création des fichiers avec une durée aléatoire
		if err := c.Write(hash[:], data, uint(rand.Intn(60))); err != nil {
			t.Fatal(err.Error())
		}
	}

}

// go test -timeout 3s -run ^TestSweep$
func TestSweep(t *testing.T) {

	c, err := cacheFile.New("/tmp/test")
	if err != nil {
		t.Fatal(err.Error())
	}

	fmt.Println(c.Sweep())
}

// go test -timeout 3s -run ^TestPurge$
func TestPurge(t *testing.T) {

	c, err := cacheFile.New("/tmp/test")
	if err != nil {
		t.Fatal(err.Error())
	}

	fmt.Println(c.Purge())
}
