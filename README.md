# Lifeofthegroup
###A work in progress telegram bot for selling your items

It accepts the bot owner's items for sale, it bothers everone that joins the channel to buy the items, accepts crypto for payment or sends the owner contact to the buyer. All of this is subject to change, of course.

#Usage
You should set the environment variables
```bash
TELEGRAM_OWNER
TELEGRAM_KEY
```
To the bot's owner username and api key, respectivelly.

Other than that it's a simple:

```bash
#on windows
go get -u github.com/cauefcr/lifeofthegroup
cd %GOPATH%\src\github.com\cauefcr\lifeoftheparty
go build
lifeofthegroup.exe

#on linux/mac
go get -u github.com/cauefcr/lifeofthegroup
cd $GOPATH/src/github.com/cauefcr/lifeoftheparty
go build
./lifeofthegroup

```
