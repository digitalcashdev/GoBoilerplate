# GoBoilerplate

A Hello-World Template for Go API Services and Proxies, and Explorers.

# Examples

- <https://rpc.digitalcash.dev>, <https://trpc.digitalcash.dev>
- <https://zmq.digitalcash.dev>, <https://tzmq.digitalcash.dev>

# Using as a Template

1. Creating a Repo using this Template:

   - <https://github.com/new?template_name=GoBoilerplate&template_owner=digitalcashdev>

2. Find and Replace the boilerplate package

   ```sh
   # sd available at https://webinstall.dev/sd

   sd 'digitalcash.dev' 'your.url' *.* */*.* */*/*.*
   sd 'DigitalCashDev' 'YourOrgUrlPath' *.* */*.* */*/*.*
   sd 'GoBoilerplate' 'YourProjectTitle' *.* */*.* */*/*.*
   sd 'goboilerplate' 'yourpackagename' *.* */*.* */*/*.*
   git mv ./cmd/goboilerplate ./cmd/yourcmdname
   ```

3. For simplicity, if you don't have `local.your.url` set up for local
   development, you may wish to switch it back to `local.digitalcash.dev`:

   ```sh
   sd 'digitalcash.dev' 'your.url' *.* */*.* */*/*.*
   ```
