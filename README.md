# gim

## Introduction

go-import-media, aka gim, is a tool for automatically importing media
from removable disks into a predefined folder structure automatically.

## Why?

We generate ~100GB of media data each week at church and currently, I am inserting one
SD card or flash drive at a time and copying to my laptop, then syncing up to the 
server and then finally syncing to the additional storage drives + backblaze. This
is super tedious and wastes a lot of time. 

The goal of this project is to automatically import and organize media when it is 
inserted into the machine. Since there are different types and sources of media,
this project needs to be able to identify the type of media and organize accordingly.

## Dependencies

Linux:
- blkid
- findmnt
- udisks2 (for mounting, unmounting, and disk poweroff without sudo access)
- udev
- systemd
- polkit

## Installation

coming soon...

## Usage

coming soon...

## License

gim is licensed under the Apache-2.0 license

Copyright (c) 2024 Steve Cross <flip@foxhollow.cc>

>  Licensed under the Apache License, Version 2.0 (the "License");
>  you may not use this file except in compliance with the License.
>  You may obtain a copy of the License at
>
>       http://www.apache.org/licenses/LICENSE-2.0
>
>  Unless required by applicable law or agreed to in writing, software
>  distributed under the License is distributed on an "AS IS" BASIS,
>  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
>  See the License for the specific language governing permissions and
>  limitations under the License.


## Links

[Project on GitHub](https://github.com/hairlesshobo/gim/)
[Project Homepage](https://www.foxhollow.cc/projects/gim/)