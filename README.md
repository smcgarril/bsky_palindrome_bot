# Bluesky Palindrome Bot

[palindrome-bot.bsky.social](https://bsky.app/profile/palindrome-bot.bsky.social)

A Bluesky bot that processes the firehose for [bsky.social](https://bsky.social/about) and checks whether new posts are palindromes.

If a palindrome is found it will post the word to its account, along with a link to the original post.

Currently the logic is limited to checking an entire post. A future implementation will check for substrings within each post (possibly using Manacher's Algorithm). Future plans also include adding a database to track statistics on palindromes to be posted on the account's profile (things like longest palindrome and the top palindromic posters).

## Quickstart

To run the bot locally, follow these steps:

(Assumes an [installation of Go](https://go.dev/doc/install))

1. Clone the repo
  ```
  $ git clone https://github.com/smcgarril/bsky_palindrome_bot.git
  $ cd bsky_palindrome_bot
  ```

2. Create a `.env` file with the following credentials:
  ```
  HANDLE=<your bsky identifier> (ex: bot-name.bsky.social)
  APIKEY=<your app password>
  ```

3. Run the bot:
  ```
  $ go run cmd/main.go
  ```

To generate an App Password, simply navigate to your bot's account in Bluesky and click `Settings` > `App Passwords`.
Do not use your personal password.

## Acknowledgements

In addition to the excellent [Bluesky](https://docs.bsky.app/docs/get-started) and [Bluesky Go SDK](https://github.com/bluesky-social/indigo) documentation, the following projects were incredibly helpful as reference:

https://github.com/wiliamvj/go-vagas/tree/post-blog
https://github.com/CharlesDardaman/blueskyfirehose
https://github.com/danrusei/gobot-bsky 