# Fortune Implementation

fortune is a minimalistic implementation of the `fortune` program we all know.
This version does not use an index file. Instead, it loads the entire fortune file into memory, parses it, and randomly selects a fortune.

### Usage

    fortune -path ./yourDirectoryFullOfFortune
    fortune -file ./fortuneCookieFile

If you don't specify a fortune cookie file path, _fortune_
defaults to using $FORTUNE_FILE or $FORTUNE_PATH if it is set.

### Fortune Cookie File Format

A fortune cookie file consists of paragraphs separated by lines containing a single '%'
character. Like this:
```
    To the world you may be just one person, but to one person, you may be the world.
    -- Brandi Snyder
    %
    Beauty is more important in computing than anywhere else in technology because software is so complicated. Beauty is the ultimate defence against complexity.
        — David Gelernter
    %
    UNIX was not designed to stop its users from doing stupid things, as that would also stop them from doing clever things.
            — Doug Gwyn
    %
    If you’re willing to restrict the flexibility of your approach, you can almost always do something better.
            — John Carmack!
```

### License
This program is licensed under the MIT-0 license. You are not required to give me credit in subsequent iterations of this program. But it will be appreciated. But I wish to know if anyone modified this program, why that is, so I'd appreciate if you sent me an email or notify me through `git`. Thanks!
