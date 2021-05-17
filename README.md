# gotorrent

Torrent client in go from scratch

## Why

Weekend project, trying to figure out how sharing with torrent works

## Features

Can join swarm of multi-file repo with magnet URL

List of support:

- BEP 003 (original BitTorrent spec)

- Random + Rarest first piece finding (outlined in the BitTorrent Economics Paper)

- BEP 009 and BEP 010 for getting metadata from other peers supporting `ut_metadata`

- Exposes port using UPnP

UI:

- none, but currently uses HAL for HATEOAS implementation https://github.com/nvellon/hal

# Acknowledgement

Project was inspired by https://blog.jse.li/posts/torrent/
