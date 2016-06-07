Usage
=====

Composer
--------

To add a new repository, for instance, the official one:

        [Composer]
            [Composer.packagist]
            Server = "https://packagist.org"
            Enabled = true

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

Npm
---

To add new repository, for instance, https://registry.npmjs.org

        [Npm]
            [Npm.npm]
            Server = "https://registry.npmjs.org"
            
Next, you need to declare the registry in npm

        npm registry set https://localhost/npm/npm

Git
---

To add new servers:

        [Git]
            [Git.github]
            Server = "github.com"
            Clone = "git@gitbub.com:{path}"
            Enabled = true
        
            [Git.drupal]
            Server = "drupal.org"
            Clone = "https://git.drupal.org/{path}"
            Enabled = true


You need to manually add git repository:

 1. Connect to the server
 2. Clone a repository
        
        git clone --mirror git@github.com:rande/gonode.git ./data/git/github.com/rande/gonode.git