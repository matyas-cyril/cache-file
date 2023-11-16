# cache-file

## Description

Générer des fichiers de cache.  
Les fichiers sont créés dans le dossier défini lors de l'initialisation.  
Le préfix par défaut des fichiers est **dsk_**  

## Fonctions
```go
// Initialiser et définir le répertoire de stockage des fichiers de cache
New(path string) (*CacheFile, error)

// Écrire les données en spécifiant une durée de validité en seconde
(c CacheFile) Write(hash []byte, data map[string][]byte, duration uint) error

// Lire le contenu d'un fichier cache
(c CacheFile) Read(fileName string) (map[string][]byte, error)

// Obtenir le nom de fichier cache correspondant au hash
(c CacheFile) GetFileName(data []byte) (string, error)

// Supprimer les fichiers de cache présents dans le répertoire.
(c CacheFile) Purge() (uint64, uint64, []error, error)

// Supprimer les fichiers invalides ou dépassés
(c CacheFile) Sweep() (uint64, uint64, []error, error)

// Obtenir le répertoire de stockage des fichiers de cache
(c CacheFile) GetPath() string

// Définir un préfixe autre que celui par défaut pour les fichiers de cache
(c *CacheFile) SetPrefix(prefix string) error

// Obtenir le préfixe utilisé pour les fichiers de cache
(c CacheFile) GetPrefix() string

// Activer le chiffrement des fichiers de cache
(c *CacheFile) EnableCrypt() bool

// Désactiver le chiffrement des fichiers de cache
(c *CacheFile) DisableCrypt() bool

// Définir manuellement une clef de chiffrement
(c *CacheFile) SetKey(key []byte) error

// Définir aléatoirement la clef de chiffrement
(c *CacheFile) SetRandomKey() []byte

// Obtenir le clef de chiffrement actuellement utilisé
(c CacheFile) GetKey() []byte
```

## Exemples

Voir **cacheFile_test.go** pour des exemples
