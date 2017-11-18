
# Go Space state button software for Techinc

Sadly, precompiled packages of Openwrt/LEDE do not support MIPS
hardware float emulation out of the box, and the (default) Go
compiler does not (yet) support software float.

Solution: compile LEDE with in-kernel hardware float emulation

## LEDE with in-kernel hardware float emulation

Follow the steps at:

https://lede-project.org/docs/guide-developer/quickstart-build-images


```bash
apt-get install subversion g++ zlib1g-dev build-essential git python rsync man-db
apt-get install libncurses5-dev gawk gettext unzip file libssl-dev wget

git clone https://git.lede-project.org/source.git lede
cd lede
git checkout -b lede-17.01

./scripts/feeds update -a
./scripts/feeds install -a

make menuconfig
# select Target System -> Atheros AR7xxx/AR9xxx
# select Target Profile -> Carambola2 board from 8Devices
make kernel_menuconfig
# select Kernel type -> [*] MIPS FPU Emulator
make

# resulting sysupgrade image:
# bin/targets/ar71xx/generic/lede-ar71xx-generic-carambola2-squashfs-sysupgrade.bin

```

## Compiling spacestate:

```bash
GOARCH=mips go build -ldflags="-s -w"

```
## Installing the spacebutton software

```bash
opkg install ca-certificates
cd /tmp
wget -O /tmp/spacestate https://<path-to-spacestate>
chmod +x /tmp/spacestate
/tmp/spacestate # test
mv /tmp/spacestate /root/spacestate
sed -i 's:^exit 0$:(/root/spacestate -key replacewithkey >/dev/null 2>\&1)\&\n&:' /etc/rc.local
```

