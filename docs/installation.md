Installation
============

Dependencies installation
-------------------------

1. Install dependencies

    apt-get install git

2. Make sure the git client have valid credentials to clone/fetch remote repositories.

    * Github: create a ssh key and add the key on github. Please note, when mirroring repository, you need
    to use the ``git@github.com:/vendor/project.git`` protocol. For instance, ``git clone --mirror git@github.com:rande/pkgmirror.git``.
        

PkgMirror installation
----------------------

1. Download the latest version from the [releases page](https://github.com/rande/pkgmirror/releases)
2. Create a configuration file ``pkgmirror.toml``

        DataDir = "/usr/local/web/pkgmirror/data"
        CacheDir = "/usr/local/web/pkgmirror/cache"
        PublicServer = "https://mirror.example.com"
        InternalServer = ":8000"

3. Create require folders

        mkdir -p /usr/local/web/{pkgmirror/data,pkgmirror/cache}

4. Start the process with a process manager

        ./pkgmirror -file pkgmirror.toml
    
If you migrate from [ekino/phpmirroring](https://github.com/ekino/php-mirroring), please read [the migration guide](migration.md)