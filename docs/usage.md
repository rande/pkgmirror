Usage
=====

Main configuration
------------------

    DataDir = "/var/lib/pkgmirror/data"
    CacheDir = "/var/lib/pkgmirror/cache"
    PublicServer = "https://mirrors.example.com"
    InternalServer = ":8000"

Composer
--------

To add a new repository, for instance, the official one:

    [Composer]
        [Composer.packagist]
        Server = "https://packagist.org"
        Enabled = true
        Icon = "https://getcomposer.org/img/logo-composer-transparent.png"

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
        Enabled = true
        Icon = "https://cldup.com/Rg6WLgqccB.svg"
            
Next, you need to declare the registry in npm

        npm set registry https://localhost/npm/npm

Git
---

To add new servers:

    [Git]
        [Git.github]
        Server = "github.com"
        Clone = "git@gitbub.com:{path}"
        Enabled = true
        Icon = "https://assets-cdn.github.com/images/modules/logos_page/GitHub-Mark.png"
    
        [Git.drupal]
        Server = "drupal.org"
        Clone = "https://git.drupal.org/{path}"
        Enabled = true
        Icon = "https://www.drupal.org/files/druplicon-small.png"


You need to manually add git repository:

 1. Connect to the server
 2. Clone a repository
        
        git clone --mirror git@github.com:rande/gonode.git ./data/git/github.com/rande/gonode.git
        
        
Bower
-----

To add a new repository, for instance, https://bower.herokuapp.com:

    [Bower]
        [Bower.bower]
        Server = "https://bower.herokuapp.com"
        Enabled = true
        Icon = "https://bower.io/img/bower-logo.svg"
        
You need to declare the mirror in your .bowerrc file:

    {
        "registry": {
            "search": ["https://localhost/bower/bower"],
            "register": "https://localhost/bower/bower"
        }
    }
