# Matrix GroupMe Go Bridge
A Matrix-GroupMe puppeting bridge

[Features & Roadmap](./ROADMAP.md)

This version is forked from the Beeper/groupme fork. This version doesn't function correctly as the `login` process doesn't work.

## Building 

### Dependencies
This requires the `libolm-dev` library.
#### Fedora

To install libolm-devel:

```
sudo dnf install libolm-devel
```

to build:
```
./build.sh
```
Then run the binary to generate the configuration file and registration files
```
./groupme -g
```

after adjusting the configuration file and configuring your matrix server to use the registration file,
run the bridge with `./groupme`

## Discussion
Matrix room: [#groupme-go-bridge:malhotra.cc](https://matrix.to/#/#groupme-go-bridge:malhotra.cc)

## Credits

Forked from https://github.com/karmanyaahm/matrix-groupme-go which was archived.
