This is an experiment for uploading large files over HTTP

## Problem statement
Whenever we need to upload large files from the browser to a server, there are some challenges from the client and from the server's point of view:

- The likelihood of an interruption in the network is bigger, therefore it's necessary to be able to continue or retry in case of failures
- Browsers and web servers have different limits on the size of data posted

## Solution
Upload files in chunks and send a final request to merge the files. We can use these chunks to retry and continue.
On the server side, each chunk will be stored in a temporary location and moved to a final location as a single file after the merge.

## Challenges
- Avoid reading the entire file during merge, to prevent allocating the whole file in memory
- FilesAPI in HTML seems a bit cumbersome
- How/When to purge chunk from temp when the upload is aborted?
- How to validate whether the file is valid after the merge?
- Provide meaningful progress feedback for the user

### Stack
- Client side: ReactJS
- Server side: Go
- Storage: file system
