package cacheFile

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Write: Écrire les données en spécifiant une durée de validité en seconde (duration)
// Si duration == 0, alors la durée du cache est infinie
func (c CacheFile) Write(hash []byte, data map[string][]byte, duration uint) error {

	// Ajout du hash
	data["_key_"] = hash

	var title string  // Définition du nom du fichier
	var exp int64 = 0 // Valeur par défaut de l'expiration
	if duration == 0 {
		title = fmt.Sprintf("%s%x0000000000", c.prefix, data["_key_"])
	} else {
		exp = time.Now().Unix() + int64(duration)
		title = fmt.Sprintf("%s%x%d", c.prefix, data["_key_"], exp)
	}

	// Ajout de l'expiration à la structure
	data["_exp_"] = []byte(strconv.FormatInt(exp, 10))

	// Init du compteur à 1
	data["_cpt_"] = []byte(strconv.FormatUint(1, 10))

	// Struct->Byte
	dataFile, err := mapToByte(data)
	if err != nil {
		return err
	}

	// Le chiffrement des données est activé
	if c.crypt {
		dataFile, err = encipher(dataFile, c.keyCrypt)
		if err != nil {
			return err
		}
	}

	// Ecriture du fichier sur le disque
	fileName := fmt.Sprintf("%s/%s", c.path, title)
	if err := writeFile(fileName, dataFile); err != nil {
		return err
	}

	return nil
}

// Read: Lire le contenu d'un fichier cache en fonction du nom du fichier
// Si le fichier cache est illisible et / ou expiré, alors il est supprimé
func (c CacheFile) Read(fileName string) (map[string][]byte, error) {

	// On utilise le Path défini
	fileName = fmt.Sprintf("%s/%s", c.path, fileName)

	file, err := readFile(fileName)
	if err != nil {
		return nil, err
	}

	// Chiffrement actif
	if c.crypt {
		file, err = decipher(file, c.keyCrypt)
		if err != nil {
			// Problème de déchiffrement
			if os.Remove(fileName) != nil {
				return nil, fmt.Errorf(fmt.Sprintf("Read: the cache file '%s' is not decryptable and could not be deleted - %s", fileName, err.Error()))
			}
			return nil, fmt.Errorf(fmt.Sprintf("Read: the cache file '%s' is not decryptable so deleted - %s", fileName, err.Error()))
		}
	}

	data, err := byteToMap(file)
	if err != nil {
		return nil, err
	}

	exp, _ := strconv.ParseInt(string(data["_exp_"]), 10, 64)

	// Le fichier cache est expiré - normalement supprimé
	if exp != 0 && exp < time.Now().Unix() {

		if os.Remove(fileName) != nil {
			return nil, fmt.Errorf(fmt.Sprintf("Read: the cache file '%s' is expired but could not be deleted", fileName))
		}

		return nil, fmt.Errorf(fmt.Sprintf("Read: the cache file '%s' is expired so deleted", fileName))
	}

	return data, nil
}

// GetFileName: Obtenir le nom de fichier cache correspondant au hash.
// S'il y a plusieurs possibilités, c'est la dernière occurance dans l'ordre
// alphabiétique, sauf si un cache sans durée est présent.
func (c CacheFile) GetFileName(data []byte) (string, error) {

	files, err := os.ReadDir(c.path)
	if err != nil {
		return "", err
	}

	re, err := regexp.Compile(fmt.Sprintf("^%s(\\d{10})$", fmt.Sprintf("%s%x", c.prefix, data)))
	if err != nil {
		return "", err
	}

	// Il est possible que plusieurs hash avec un epoc différent soit présent. Un Purge ou Sweep peut résoudre le pb
	// Donc on récupére une liste et la dernière entrée est forcément de dernier créé
	listFiles := []string{}

	for _, file := range files {
		if !file.IsDir() && re.MatchString(file.Name()) {
			listFiles = append(listFiles, file.Name())
			// Cache sans timeout
			if file.Name() == fmt.Sprintf("%s%x%s", c.prefix, data[:], "0000000000") {
				return file.Name(), nil
			}
		}
	}

	length := len(listFiles)
	if length == 0 {
		return "", fmt.Errorf("GetFileName: no file cache available")
	}

	return listFiles[length-1], nil
}

// Purge: Supprimer les fichiers présents dans le répertoire.
// Les fichiers doivent correspondre au pattern
// Il n'y a de contrôle. Même si le fichier cache est valide, il sera supprimé
func (c CacheFile) Purge() (uint64, uint64, []error, error) {

	arrayErr := []error{}

	files, err := os.ReadDir(c.path)
	if err != nil {
		return 0, 0, arrayErr, err
	}

	// Définition du pattern des fichiers. Prise en compte des fichiers .PREFIX s'ils existent
	re, err := regexp.Compile(fmt.Sprintf("^\\.{0,1}%s\\S{128}\\d{10}$", c.prefix))
	if err != nil {
		return 0, 0, arrayErr, err
	}

	var ok, ko uint64
	for _, file := range files {
		if !file.IsDir() && re.MatchString(file.Name()) {
			fileName := fmt.Sprintf("%s/%s", c.path, file.Name())
			if err := os.Remove(fileName); err != nil && !strings.Contains(err.Error(), ": no such file or directory") {
				arrayErr = append(arrayErr, err)
				ko++
				continue
			}
			ok++
		}
	}

	if len(arrayErr) > 0 {
		return ok, ko, arrayErr, fmt.Errorf("purge: errors during command execution")
	}

	return ok, ko, nil, nil
}

// Sweep: Supprimer les fichiers invalides ou dépassés
// Dans un premier temps la suppression ne s'effectue que par le nom, puis
// par une analyse plus détaillée (lecture du fichier)
func (c CacheFile) Sweep() (uint64, uint64, []error, error) {

	arrayErr := []error{}

	files, err := os.ReadDir(c.path)
	if err != nil {
		return 0, 0, arrayErr, err
	}

	// Définition du pattern des fichiers
	re, err := regexp.Compile(fmt.Sprintf("^%s(\\S{128})(\\d{10})$", c.prefix))
	if err != nil {
		return 0, 0, arrayErr, err
	}

	var ok, ko uint64
	for _, file := range files {

		d := re.FindStringSubmatch(file.Name())

		if !file.IsDir() && len(d) == 3 {

			hash := string(d[1])
			exp, _ := strconv.ParseInt(string(d[2]), 10, 64)
			fileName := fmt.Sprintf("%s/%s", c.path, file.Name())

			// On se base sur le nom du fichier uniquement
			if exp != 0 && exp < time.Now().Unix() {

				if err := os.Remove(fileName); err != nil && !strings.Contains(err.Error(), ": no such file or directory") {
					arrayErr = append(arrayErr, err)
					ko++
					continue
				}
				ok++
				continue
			}

			// Analyse du fichier donc plus long
			data, err := c.Read(file.Name())
			if err != nil && !strings.Contains(err.Error(), ": no such file or directory") {

				// Problème au chargement du fichier cache
				if err := os.Remove(fileName); err != nil && !strings.Contains(err.Error(), ": no such file or directory") {
					arrayErr = append(arrayErr, err)
					ko++
					continue
				}
				ok++
				continue
			}

			// Incohérence entre hash nom de fichier et valeur interne ou problème de timeout
			dataExp, errExp := strconv.ParseInt(string(data["_exp_"]), 10, 64)
			if !(fmt.Sprintf("%x", data["_key_"]) == hash && errExp == nil && exp == dataExp && (exp == 0 || dataExp >= time.Now().Unix())) {

				if err := os.Remove(fileName); err != nil && !strings.Contains(err.Error(), ": no such file or directory") {
					arrayErr = append(arrayErr, err)
					ko++
					continue
				}
				ok++
				continue

			}

		}

	}

	if len(arrayErr) > 0 {
		return ok, ko, arrayErr, fmt.Errorf("sweep: errors during command execution")
	}

	return ok, ko, nil, nil
}
