## Banter Bus


### CI

At the moment we don't promote images from dev to prod. We just rebuild the images on main branch, as the Nix
docker build is immutable. We could potentially solve this to save some time to deploy, by re-tagging a dev image.
Then we would need to tag them using a commit. By since we squash and merge we lose the commit. So we need to use
the Gitlab API to fetch commit from the MR. Then get the latest commit from that MR, use that to get an image to
retag.
