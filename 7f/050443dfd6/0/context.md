# Session Context

## User Prompts

### Prompt 1

Apple IDを契約しましたこの後どのような手順でリリースできるようになりますか？

### Prompt 2

Step 1 から順番に手順を教えてください。

### Prompt 3

% codesign -dv --verbose=4 $(which meow)
Executable=/Users/135yshr/go/bin/meow
Identifier=a.out
Format=Mach-O thin (arm64)
CodeDirectory v=20400 size=33438 flags=0x20002(adhoc,linker-signed) hashes=1042+0 location=embedded
VersionPlatform=1
VersionMin=786432
VersionSDK=786432
Hash type=sha256 size=32
CandidateCDHash sha256=5c89776f7a401621d43612fad48cc4d3e90e892b
CandidateCDHashFull sha256=5c89776f7a401621d43612fad48cc4d3e90e892bbf917aaf5a4abbe631a17310
Hash choices=sha256
CMSDigest=5c89776f7...

### Prompt 4

% codesign -dv --verbose=4 $(which meow)
Executable=/opt/homebrew/Caskroom/meow/0.2.0/meow
Identifier=meow
Format=Mach-O thin (arm64)
CodeDirectory v=20500 size=26629 flags=0x10000(runtime) hashes=827+2 location=embedded
VersionPlatform=1
VersionMin=786432
VersionSDK=786432
Hash type=sha256 size=32
CandidateCDHash sha256=4f0bd60f0846bed80b7276c38f1e634fa5d441c8
CandidateCDHashFull sha256=4f0bd60f0846bed80b7276c38f1e634fa5d441c88e37ded8e96d0ec7c812e233
Hash choices=sha256
CMSDigest=4f0bd60f084...

### Prompt 5

<bash-input>git co main && git pull</bash-input>

### Prompt 6

<bash-stdout>Switched to branch 'main'
Your branch is up to date with 'origin/main'.
From github.com:135yshr/meow
   a57d09d..f4fb780  main       -> origin/main
 * [new tag]         v0.2.0     -> v0.2.0
Updating a57d09d..f4fb780
Fast-forward
 .github/workflows/goreleaser.yaml    |  5 +++++
 .github/workflows/slash-release.yaml |  5 +++++
 .goreleaser.yaml                     | 11 +++++++++++
 3 files changed, 21 insertions(+)</bash-stdout><bash-stderr></bash-stderr>

### Prompt 7

このリポジトリの状態を確認してください。
オープンソースとして不足している情報があれば教えてください。
他の有名なオープンソースを確認した上で提案してください。

### Prompt 8

はい。作成してください。
なのであれば、全て作成してください。

１点補足があります。
.github/PULL_REQUEST_TEMPLATE.md は、~/projects/135yshr/documents/articles/499cd6335b5fa6.md no

### Prompt 9

[Request interrupted by user]

### Prompt 10

はい。作成してください。
なのであれば、全て作成してください。

１点補足があります。
.github/PULL_REQUEST_TEMPLATE.md は、~/projects/135yshr/documents/articles/499cd6335b5fa6.md のルールを追加してください

### Prompt 11

はい。作成してください。
なのであれば、全て作成してください。

１点補足があります。
.github/PULL_REQUEST_TEMPLATE.md は、~/projects/135yshr/documents/articles/499cd6335b5fa6.md のルールを追加してください

