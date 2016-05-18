package pkgmirror

import (
	"github.com/boltdb/bolt"
	"fmt"
	"strings"
	log "github.com/Sirupsen/logrus"
	"sync"
	"encoding/json"
)

type PackageInformation struct {
	Url           string
	PackageResult PackageResult
	Package       string
	Exist         bool
}

func PackagistMirror() {
	db, _ := LoadDB("packagist")

	baseServer := "https://packagist.org"

	bName := []byte("data")

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bName)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		return nil
	})

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("data"))

		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			fmt.Printf("%s\n", k)
		}

		return nil
	})

	dm := NewDownloadManager()

	ps := &PackagesResult{}

	url := fmt.Sprintf("%s/packages.json", baseServer)

	log.WithFields(log.Fields{
		"handler": "packagist",
		"server": baseServer,
		"url": url,
	}).Info("Loading packages.json")

	if err := LoadRemoteStruct(url, ps); err != nil {
		log.WithFields(log.Fields{
			"handler": "packagist",
			"server": baseServer,
			"url": url,
			"error": err.Error(),
		}).Error("Error loading packages.json")
	} else {
		log.WithFields(log.Fields{
			"handler": "packagist",
			"server": baseServer,
			"url": url,
		}).Info("End loading packages.json")
	}

	var wg sync.WaitGroup
	var lock sync.Mutex

	PackageListener := make(chan PackageInformation)

	go dm.Wait(10, func(id int, done chan <- struct{}, pkgs <-chan PackageInformation) {
		for pkg := range pkgs {
			p := &PackageResult{}

			log.WithFields(log.Fields{
				"handler": "packagist",
				"server": baseServer,
				"url": pkg.Url,
				"package": pkg.Package,
				"id": id,
			}).Info("Loading package information")

			if err := LoadRemoteStruct(pkg.Url, p); err != nil {
				log.WithFields(log.Fields{
					"handler": "packagist",
					"server": baseServer,
					"url": pkg.Url,
					"package": pkg.Package,
					"error": err.Error(),
				}).Error("Error loading package information")
			}

			pkg.PackageResult = *p

			PackageListener <- pkg

			wg.Done()
		}
	})

	go func() {
		cpt := 0

		for {
			select {
			case pkg := <-PackageListener:

				cpt++

				data, _ := json.Marshal(pkg.PackageResult)

				lock.Lock()

				err := db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte("data"))
					err := b.Put([]byte(pkg.Url), data)

					if err != nil {
						log.WithFields(log.Fields{
							"handler": "packagist",
							"url": pkg.Url,
							"package": pkg.Package,
							"error": err.Error(),
						}).Error("Bolt error")
					}

					return nil
				})

				lock.Unlock()

				if err != nil {
					log.WithFields(log.Fields{
						"handler": "packagist",
						"package": pkg.Package,
						"url": pkg.Url,
						"error": err.Error(),
					}).Error("Error updating/creating bolt entry")
				} else {
					log.WithFields(log.Fields{
						"handler": "packagist",
						"package": pkg.Package,
						"url": pkg.Url,
						"counter": cpt,
					}).Info("Package information saved!!")
				}
			}
		}
	}()

	for name, sha := range ps.ProviderIncludes {
		url := fmt.Sprintf("%s/%s", baseServer, strings.Replace(name, "%hash%", sha.Sha256, -1))

		log.WithFields(log.Fields{
			"handler": "packagist",
			"server": baseServer,
			"url": url,
			"provider": name,
		}).Info("Loading provider information")

		pr := &ProvidersResult{}

		if err := LoadRemoteStruct(url, pr); err != nil {
			log.WithFields(log.Fields{
				"handler": "packagist",
				"server": baseServer,
				"url": url,
				"provider": name,
				"error": err.Error(),
			}).Error("Error loading provider information")
		}

		for name, sha := range pr.Providers {
			url := fmt.Sprintf("%s/p/%s$%s.json", baseServer, name, sha.Sha256)

			p := PackageInformation{
				Url: url,
				Package: name,
				Exist: false,
			}

			lock.Lock()
			db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("data"))
				v := b.Get([]byte(url))

				p.Exist = len(v) != 0

				return nil
			})
			lock.Unlock()

			if !p.Exist {
				log.WithFields(log.Fields{
					"handler": "packagist",
					"server": baseServer,
					"package": name,
					"url": url,
				}).Info("Add new package")

				wg.Add(1)

				dm.Add <- p
			} else {
				log.WithFields(log.Fields{
					"handler": "packagist",
					"server": baseServer,
					"package": name,
					"url": url,
				}).Info("Skipping package")
			}
		}
	}

	log.WithFields(log.Fields{
		"handler": "packagist",
	}).Info("Wait for goroutines to complete")

	wg.Wait()
}