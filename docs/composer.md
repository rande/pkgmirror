Composer mirroring
==================

The mirroring will fetch the latest information from packagist.org and keep the data in a local storage (boltdb). The
archive paths can also be updated with the local git mirroring.

Composer workflow
-----------------

1. Load https://packagist.org/packages.json, which contains definitions to some providers. 

        provider-includes: {
            p/provider-2013$%hash%.json: {
                sha256: "81839b9e7c94fdecc520e0e33f8e47b092079568ccfa319650db0e353412bfc3"
            },
            p/provider-2014$%hash%.json: {
                sha256: "27fb04c654fb35ac2cb50183cc03861396cdacfc57e5ce94735e71a44a393bc4"
            },
            p/provider-2015$%hash%.json: {
                sha256: "9c5310ed37ea7fd7243e26a62b150f0c7c257065236a208301712f524a6e68e9"
            },
            p/provider-2015-07$%hash%.json: {
                sha256: "c2b3d17ececc1cab2cdf039c5016513c51e83f6d0998ebf4ee4d37eb401f5b4d"
            },
            p/provider-2015-10$%hash%.json: {
                sha256: "898cee210f3b6ee1b8ae98ac4875dbe9632c480b66ccc7382b21ff75ca2fad5a"
            },
            p/provider-2016-01$%hash%.json: {
                sha256: "a2ebcb11c730eeb56c42af336e6654a2b58f7c62e9edaa3528fa4ad6def8bafe"
            },
            p/provider-2016-04$%hash%.json: {
                sha256: "2dd634aa1adabfb1c82e6f93ac65bb35a120f651e6dd139787e3e09521427467"
            },
            p/provider-archived$%hash%.json: {
                sha256: "dcf5bde9f42034979a3475a33d0282b60ce8db28c4b1ab98728a6c7b8c467e00"
            },
            p/provider-latest$%hash%.json: {
                sha256: "09fc55f7e0e166e7a96d9d07460f87b88fd1aa78ba8bd4454c3c9e953d7e3253"
            }
        }

    The sha256 value is the hash of the target file. 

2. Load the provider files, each file contains references to package file.

        {
            "providers": {
                "0s1r1s\/dev-shortcuts-bundle": {
                    "sha256": "6c7710a1ca26d3c0f9dfc4c34bc3d6e71ed88d8783847ed82079601401e29b18"
                },
                "0x20h\/monoconf": {
                    "sha256": "9515a0ee8fce44be80ed35292384c2f908cabbf6a710099f4743b710bc47607e"
                },
                "11ya\/excelbundle": {
                    "sha256": "65dccb7f2d57c09c19519c1b3cdf7cbace1dfbf46f43736c2afcb95658d9c0f1"
                },
                ....
        }
        
3. Load the package information.

Mirroring workflow
------------------

1. Load the packages.json file
2. Iterate over the providers.
4. Download the package definition and 
    - alter path if required, 
    - compute new hash
    - store the package in data layer using bzip compression to save bandwidth and local storage.
5. Update packages.json and providers.json to use the new hash and recompute the final hash for each provider.
6. Clean old references

Storage
-------

The storage layer uses [boltdb](https://github.com/boltdb/bolt). A packagist.org mirror is about 512MB on disk, 
the file is located in the ``PATH/composer/packagist.db``. 
