# keep 📔

🌟 Keep notes with ease 🌟

Keep is a CLI tool that helps you to keep short notes. Simple as that.

## Usage 🔍

Below are some example usages for keep.

Create note:
```sh
keep "this is a note"
```

Keep allows you to create groups to organize notes. To create a group, you must
give a title and a description as follow:
```sh
keep group "books" "Here are some books that I want to read"
```

To store a note in a group you can do as follow:
```sh
keep "books" "Programming Language Pragmatics"
```

To read all notes from a group do:
```sh
keep read books
```

If you want to list all groups you've created:
```sh
keep list
```

## Installation

Download a build from download page here in github. After that, the installation
depends on your operating system. 

For linux:
1 - Move the app binary into `/usr/bin`
```sh
sudo mv ~/Downloads/keep /usr/bin/keep
```

2 - Give execution permission for the binary:
```sh
chmod -x /usr/bin/keep
```

3 - Restart your terminal session.
4 - Enjoy!