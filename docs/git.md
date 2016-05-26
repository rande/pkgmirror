Git Mirroring
=============

The git mirroring service allows to mirror git repositories into a local server.

In order to mirror a repository, go to the working dir of the running binary and start the following command: 
``git clone --mirror https://github.com/rande/gonode.git ./data/git/rande/gonode.git``. This will mirror the repository.
Please note, the mirroring solution use a [dump http server](https://git-scm.com/book/tr/v2/Git-on-the-Server-The-Protocols) so 
the cloning operation will only works if ``update-server-info`` command is run. The binary will run this command automatically.

- git/hostname/vendor/package.git/ref.zip : return a zip file using zip archive with the local mirror available in ./data/git. 
  for instance: ``curl  localhost:8000/git/github.com/rande/gonode.git/master.zip > master.zip``
   