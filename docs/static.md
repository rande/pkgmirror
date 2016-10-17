Static mirroring
================

This service is a simple proxy to remote file. Each server need to be configured.

For now, there is no specific configuration, the current behavior is:
 
 - if ``statusCode == 200`` the file will be stored into the local cache
 - if ``statusCode == 404`` the file will no be stored and a 404 code will be send to the suer
 - if ``statusCode == 302`` the internal http lib will follow redirect and the final will be stored with the initial provided path.
 - any other code will result of an "Internal Server Error"