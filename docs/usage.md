Usage
=====

Composer
--------

To add a new repository, for instance, the official one:

        [Composer]
            [Composer.packagist]
            Server = "https://packagist.org"

Next, you need to declare the mirror in your ``composer.json`` file:

        {
            "repositories":[
                { "packagist": false },
                { "type": "composer", "url": "https://localhost/composer/packagist"}
            ],
        
            "require": {
                "sonata-project/exporter": "*"
            }
        }

The ``packagist`` key is used here as an example.

Git
---

You need to manually add git repository:

 1. Connect to the server
 2. Clone a repository
        
        git clone --mirror git@github.com:rande/gonode.git ./data/git/github.com/rande/gonode.git