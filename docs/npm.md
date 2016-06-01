NPM mirroring
=============


Mirroring Workflow
------------------

1. Load https://registry.npmjs.org/-/all json. It is a 130MB json file with partial informations.

        {
            "_updated":1464585613977,
            "package_name": {  
                 "name":"bower",
                 "description":"The browser package manager",
                 "dist-tags":{  
                     "latest":"1.7.9",
                     "beta":"1.7.9"
                 },
                 "maintainers":[{  
                     "name":"desandro",
                     "email":"desandrocodes@gmail.com"
                 }],
                 "author":{  
                     "name":"Twitter"
                 },
                 "users":{  
                     "sjonnet":true,
                     "sjonnet19":true,
                     "vincentmac":true,
                     "...": true,
                 },
                 "repository":{  
                     "type":"git",
                     "url":"git+https:\/\/github.com\/bower\/bower.git"
                 },
                 "homepage":"http:\/\/bower.io",
                 "bugs":{  
                     "url":"https:\/\/github.com\/bower\/bower\/issues"
                 },
                 "readmeFilename":"README.md",
                 "keywords":[  
                     "bower"
                 ],
                 "license":"MIT",
                 "time":{  
                     "modified":"2016-04-05T11:54:07.456Z"
                 },
                 "versions": {  
                     "1.7.9":"beta"
                 }
            },
        }
        
2. Update the local metadata with the modified date, if changed then update related package reference.
3. The package update will download the package information from ``https://registry.npmjs.org/package_name`` and update tarbal reference to point to the local entry point.


Entry Points
------------

* Get package information: ``/npm/package_name``
* Download archive: ``/npm/package_name/-/package_name-version.tgz``
