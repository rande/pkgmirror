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

        [Composer.drupal8]
        Server = "https://packages.drupal.org/8"
        Enabled = true
        Icon = "https://www.drupal.org/files/druplicon-small.png"

        [Composer.drupal7]
        Server = "https://packages.drupal.org/7"
        Enabled = true
        Icon = "https://www.drupal.org/files/druplicon-small.png"


Next, you need to declare the mirror in your ``composer.json`` file:

    {
        "repositories":[
            { "packagist": false },
            { "type": "composer", "url": "https://localhost/composer/packagist"}
            { "type": "composer", "url": "https://localhost/composer/drupal8"}
        ],
        "require": {
            "sonata-project/exporter": "*"
        }
    }

The ``packagist`` key is used here as an example.


> You also need to setup `git` and `static` configuration to be able to download assets or clone repository.

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


If the ``Clone`` settings is not set, you need to manually add git repository:

 1. Connect to the server
 2. Clone a repository
        
        git clone --mirror git@github.com:rande/gonode.git ./data/git/github.com/rande/gonode.git
        
        
Bower
-----

To add a new repository, for instance, https://registry.bower.io:

    [Bower]
        [Bower.bower]
        Server = "https://registry.bower.io"
        Enabled = true
        Icon = "https://bower.io/img/bower-logo.svg"
        
You need to declare the mirror in your .bowerrc file:

    {
        "registry": {
            "search": ["https://localhost/bower/bower"],
            "register": "https://localhost/bower/bower"
        }
    }

Static
------

To add a new server:

    [Static]
        [Static.drupal]
        Server = "https://ftp.drupal.org/files/projects"
        Icon = "https://www.drupal.org/files/druplicon-small.png"
        
You can now download file from ``https://localhos/static/drupal/panopoly-7.x-1.40-core.tar.gz``

