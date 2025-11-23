# Instructions for generating test data

```
skopeo copy docker-daemon:docker.io/tianon/toybox:0.8.11 oci-archive:toybox.tar:docker.io/tianon/toybox:0.8.11
mkdir -p oci
tar -C oci -xf toybox.tar
sudo umoci unpack --image oci:docker.io/tianon/toybox:0.8.11 unpacked
sudo mkfs.erofs toybox.img unpacked/rootfs/
```
