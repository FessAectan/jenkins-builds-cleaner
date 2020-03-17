# jenkins-builds-cleaner
A simple tool for deleting Jenkins' builds.    
Usually, Jenkins' jobs create many files on the underlying filesystem and they can overflow filesystem (storage or inodes).  
jenkins-builds-cleaner deletes all build in every Jenkins' job except 10 last builds.  
Run it on midnight and keep calm.   

