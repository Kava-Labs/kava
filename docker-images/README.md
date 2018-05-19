This Dockerfile is to build cosmos-sdk. It needs to be in the cosmos-sdk repo to build.

It modifies the existing Dockerfile in the cosmos-sdk:

 - split up commands to make use of layers to make rebuilds and uploads faster
 - switch `ENTRYPOINT gaiad` to `CMD gaiad` to save typing `--entrypoint` all the time