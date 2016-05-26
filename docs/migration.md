Migration from ekino/phpmirroring
=================================

The project ``ekino/php-mirroring`` is deprecated, the current page explains how to migrate your project from ``ekino/php-mirroring`` to ``pkgmirror``.

Archive Update
--------------

The old urls look likes: ``http://oldserver.com/cache.php/github.com/doctrine/cache/47cdc76ceb95cc591d9c79a36dc3794975b5d136.zip``

The new urls now are: ``http://newserver.com/git/github.com/doctrine/cache/47cdc76ceb95cc591d9c79a36dc3794975b5d136.zip``


Repository Update
-----------------

The old urls look likes: ``git@oldserver.com:/mirrors/github.com/doctrine/cache.git``

The new urls now are: ``http://newserver.com/git/github.com/doctrine/cache.git``


Migration
---------

You can update your ``composer.lock`` file by running the command : ``cat packages.lock | sed -e 's|http://oldserver.com/cache.php|https://newserver.com|' | sed -e 's|git@oldserver.com:/mirrors|https://newserver.com/git|'``.

If the output is ok, you can redirect the output to the ``composer.lock`` file

You also need to update the ``composer.json`` file:

Before:

    "repositories":[
        { "packagist": false },
        { "type": "composer", "url": "http://oldserver.com"}
    ],

After:

    "repositories":[
        { "packagist": false },
        { "type": "composer", "url": "http://newserver.com/packagist"}
    ],

If you use the same domain name, you can clear local composer cache to avoid any issues: ``rm -rf ~/.composer/cache``
