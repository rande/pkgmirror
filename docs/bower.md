Bower mirroring
=============

Mirroring Workflow
------------------

1. Load https://bower.herokuapp.com/packages json. It is a 4.3MB json file with all informations.

        [
            {
                "name": "10digit-geo",
                "url": "https://github.com/10digit/geo.git"
            },
            {
                "name": "10digit-invoices",
                "url": "https://github.com/10digit/invoices.git"
            },
            {
                "name": "10digit-legal",
                "url": "https://github.com/10digit/legal.git"
            }
        
2. Update the local metadata.


Entry Points
------------

* Get package information: ``/bower/bower/packages/package_name``
* Download all packages: ``/bower/bower/packages``
