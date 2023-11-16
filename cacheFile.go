package cacheFile

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type CacheFile struct {
	path     string // Répertoire où sont stockés les fichiers
	prefix   string // Prefix du nom des fichiers
	keyCrypt []byte // clef de chiffrement
	crypt    bool   // Activer / Désactiver le chiffrement des fichiers
}

// New:
func New(path string) (*CacheFile, error) {

	path = strings.TrimSpace(path)
	if len(path) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("path name '%s' is not valid", path))
	}

	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf(fmt.Sprintf("path %s is not absolute", path))
	}

	// Si le dossier n'existe pas, alors on va le créer !!!!
	if isPathExist(path) != nil {

		if err := makeDirectory(path); err != nil {
			return nil, err
		}

	}

	// Peut-on écrire dans le répertoire
	if err := isWritable(path); err != nil {
		return nil, err
	}

	return &CacheFile{
		path:   path,
		prefix: "dsk_",
	}, nil

}

// GetPath: Obtenir le path actuel
func (c CacheFile) GetPath() string {
	return c.path
}

// SetPrefix: Définir le préfixe des noms des fichiers de cache
// La longueur min est 1 et max 50
// Les caractères autorisés sont : A-Z a-z 0-9
func (c *CacheFile) SetPrefix(prefix string) error {
	re := regexp.MustCompile(`^[A-Za-z0-9]{3,49}\_$`)
	if re.MatchString(prefix) {
		c.prefix = prefix
	}
	return fmt.Errorf(fmt.Sprintf("set '%s' prefix invalid", prefix))
}

// GetPath: Obtenir le prefix des fichiers actuel
func (c CacheFile) GetPrefix() string {
	return c.prefix
}

// EnableCrypt: Activer le chiffrement
// Si la clef de chiffrement est vide, une clef sera générée aléatoirement
// Retourne l'état de la clef
func (c *CacheFile) EnableCrypt() bool {
	if len(c.keyCrypt) == 0 {
		c.crypt = false
	} else {
		c.crypt = true
	}
	return c.crypt
}

// DisableCrypt: Désactiver le chiffrement et retourner l'état de la clef
func (c *CacheFile) DisableCrypt() bool {
	c.crypt = false
	return c.crypt
}

// GetStatusCrypt: Obtenir le status Actif / Non actif du chiffrement
func (c CacheFile) IsCrypt() bool {
	return c.crypt
}

// SetKey: Clef de chiffrement défini par l'utilisateur.
func (c *CacheFile) SetKey(key []byte) (_err error) {

	defer func() {
		if err := recover(); err != nil {
			_err = fmt.Errorf(_err.Error())
		}
	}()

	if len(key) == 0 {
		c.keyCrypt = []byte{}
		c.crypt = false
	} else {
		c.keyCrypt = genAESKey(key)
	}
	return nil
}

// SetRandomKey: Définir aléatoirement la clef de chiffrement
// Retourne la clef générée
func (c *CacheFile) SetRandomKey() []byte {
	c.keyCrypt = randKey(3)
	if len(c.keyCrypt) == 0 {
		c.crypt = false
	}
	return c.keyCrypt
}

// GetKey: Obtenir la clef de chiffrement
func (c CacheFile) GetKey() []byte {
	return c.keyCrypt
}
