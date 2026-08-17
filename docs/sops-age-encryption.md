# SOPS Encryption with Age

[SOPS](https://github.com/getsops/sops) is an editor that handles encryption and decryption of configuration files. It supports `yaml`, `json`, `ini`, and `env` formats, and pairs with several key management systems (KMS) [backends](https://getsops.io/docs/usage/identities/) to facilitate the encryption/decryption process.

[Age](https://github.com/FiloSottile/age) is a vendor agnostic, light weight, asymmetrical file encryption tool. It's often recommended over PGP for being more modern and easier to use. The combination of [SOPS + Age](https://getsops.io/docs/usage/identities/age/) enables a secure way to encrypt configuration secrets and safely commit them to a git repository.

This project uses Age to encrypt the contents of a `.env` file which contains environment variable values used by our [Pulumi](pulumi-iac.md) applications. They are prefixed with `ECA` to distinguish them and avoid collision with other programs, because for example, `${AWS_ACCESS_KEY_ID}` and `${AWS_SECRET_ACCESS_KEY}` are used by pretty much anything that interacts with Object Storage via the AWS SDK. The table below outlines those Edgecase Apparel environment variables which are sourced from `.env`.

| Key                      |  Value                      |
| ------------------------ | --------------------------- |
| ECA_PULUMI_CONFIG_SECRET | ${PULUMI_CONFIG_PASSPHRASE} |
| ECA_OBJ_ACCESS_KEY       | ${AWS_ACCESS_KEY_ID}        |
| ECA_OBJ_SECRET_KEY       | ${AWS_SECRET_ACCESS_KEY}    |
| ECA_LINODE_TOKEN         | ${LINODE_TOKEN}             |

> [!NOTE]
> You can also just put the `ECA` environment variables in your shell profile (`~/.bashrc`, `~/.zshrc`):
>
> ```bash
> export ECA_PULUMI_CONFIG_SECRET=<PULUMI_CONFIG_PASSPHRASE>
> export ECA_OBJ_ACCESS_KEY=<OBJ_ACCESS_KEY>
> export ECA_OBJ_SECRET_KEY=<OBJ_SECRET_KEY>
> export ECA_LINODE_TOKEN=<LINODE_API_TOKEN>
> ```

## Getting Started

Once we have shared the **private** Age key through other means, such as a shared vault in a password manager, the value of this secret key needs to exported. In other words, you'll want to add this to your shell profile as well.

```bash
export SOPS_AGE_KEY=<AGE_SECRET_KEY>
```

If not set, then SOPS will search the filesystem for a keyfile. If you have access to the keyfile―whether shared by a team member, or from generating your own Age keys―the default location for it is `$HOME/.config/sops/age/`.

```bash
mkdir -p ~/.config/sops/age
mv key.txt ~/.config/sops/age/key.txt
```

The _environment variable_ method should work just fine for now, since our team is only sharing _one_ Age key―it's unlikely we'll need to distribute the the keyfile. That said however, it's  relatively just as simple to generate _individual_ Age keys if needed. Each person just runs `age-keygen` on their own machine, and then it's the same as above concerning whether to use an environment variable or keyfile.

```bash
age-keygen -o key.txt
```

Lastly, update the shared `.sops.yaml` file to include public keys of all recipients. See the section on using [multiple encryption keys](#multiple-encryption-keys) for more on that.

##  Decrypt the ENV
Suppose a team member is getting setup for the very first time, to which the first steps include cloning this repo and activating the Flox environment.

```bash
git clone https://github.com/akamai-developers/edgecase-apparel.git
cd edgecase-apparel
flox activate
```

At this point the user has pulled down the `.secret.env.enc` file with the rest of the repo. Assuming we shared the **private** Age key with them already, they just need to simply export its value as `$SOPS_AGE_KEY`, and then decrypt. After sourcing the resulting `.env` file, all of the `ECA` variables have been set. See how easy that was!

```bash
sops -d .secret.env.enc > .env
source .env
```

> [!NOTE]
> This may be an obvious reminder, but we'll need to do this again anytime someone has committed an updated `.secret.env.enc` file the repo, such as after a round of secrets rotation.

## Create and Edit Secrets

Say we need to add or update the Linode API token. Just run `sops` with `.secret.env.enc` as the only argument. This will decrypt and open it in your `$EDITOR`, allowing you to make the changes you need as _plaintext_.

```bash
ECA_PULUMI_CONFIG_SECRET=<OLD_VALUE>
ECA_OBJ_ACCESS_KEY=<OLD_VALUE>
ECA_OBJ_SECRET_KEY=<OLD_VALUE>
ECA_LINODE_TOKEN=<NEW_VALUE>  # new Linode API token value
```

Make your modifications...then save and quit! SOPS will automatically encrypt it on the way out, using your **public** Age key. It does this based on `creation_rules` defined in the `.sops.yaml` configuration file.

```yaml
creation_rules:
  - path_regex: \.secret\.env\.enc$
    age: age1jdfjhhdce2wqmgpngc5cx04ljwtnlnn9033p6gkfvmm9fv5ydsastnk8p9
```

As you can see, it's possible to define [multiple creation rules](https://getsops.io/docs/usage/identities/config-file/). Our example only has one at the moment, where `path_regex` specifies `.secret.env.enc` as the output file, and `age` specifies the **public** key to encrypt it with. The snippet below is an example of what the final encrypted file looks like.

```json
{
	"data": "ENC[AES256_GCM,data:+zKX0dPSSkvRwVA8Q+x17qp9cLW+BvstHzqhTu4rvzmptM+QTeQy3eeKKwSSE3Ugm1M+l9fe9CrZiqzWsMOdL+1QEOrFa/eMC9BYw+Y93FBOx7Up6UOJsTsVvjJWLDgQJdSlNhNHY2AZlrNjTly56hzelEWCwVriK2WKqLTs/aJa1ZOqGAXwg1o6spKg4/15NviMeJSBxPvLV6XhDKqADY3v5I6ZlWkVcDBVpkQ7znb7ePi1ubjoPVsPCf0+S5XH7ZPyl+L8AOxsibIQPwgbj7sXOUlkehUozEwa9iusGuAOTug6g/zUK6GZww+aePaqOoS/jajapL1m,iv:B6vj3/HCZlMS8mcAWc45m5KQlPaio7UzeAzp0T0qB90=,tag:YBCHdINwcCngaSJdhGKHJA==,type:str]",
	"sops": {
		"age": [
			{
				"recipient": "age1jdfjhhdce2wqmgpngc5cx04ljwtnlnn9033p6gkfvmm9fv5ydsastnk8p9",
				"enc": "-----BEGIN AGE ENCRYPTED FILE-----\nYWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSBkMk9JTEFBejVVbUZraGFz\nNm1XUE1wemdDUytsQXJtS201OHVqNkxXVEI4CmxZd0JkbFcwL0ttRHdiMmF6RElD\nQWZOd2NObzZ2SEhQSDVYTVRrMkNXTncKLS0tIDNFNHJTMjRYT2pIU0VsRWl0b1dN\neGlMVFA1Q0VxWUtyUFBFaHYrQjdva2sK+k2En020tPDHzrkVgnPpSkYXLzddLh3y\nSoKcMjPJA9UJLdn9IFTZsl921/LoqBGs5AxYi5URV6c4U9MYNT1FSA==\n-----END AGE ENCRYPTED FILE-----\n"
			}
		],
		"lastmodified": "2026-08-13T19:05:56Z",
		"mac": "ENC[AES256_GCM,data:Rx3E9AhYGhN6OOsq9As4zCR7KxKSo9wM9zAcCz6IcAFrLIgRSjKRlHqNiMNiKzrOkUSvpvq2NalNOh2Hm5w4+P6R9EPLUA5W1yH7kj8dTxO7KnoDXHKLPTfhcPte4u8PqbGutHHAjv/uarBeeTEvC5YJO0bi/oq08XLIontDieQ=,iv:LQgK2fMQpUG1Iz0LRLNRPaC8y9vCBpVPl6DI85ZpKAU=,tag:CGV5VDO9PDrBLrHPChtHjQ==,type:str]",
		"version": "3.12.2"
	}
}
```

> [!NOTE]
> A common naming pattern for SOPS files is something like `secrets.enc.yaml`, whereas for our project, the naming is `.secret.env.enc`. We still adheres to the use of double extensions, to clearly distinguish it from dotfiles we _don't_ intend to commit to the repo. Our deviation is just to avoid causing `yaml` or `json` _unmarshal errors_ due to the fact it's not a `yaml` or `json` file that we're trying to encrypt! It's an `env` file, so it's rightfully formatted that way. The core reason for swapping the naming order of these double extensions (`env` and `enc`), is to safeguard against accidental glob matching in the `.gitignore` file. The keys also get encrypted as a bonus.

## Multiple Encryption Keys

To use multiple Age keys, allowing multiple _recipients_ to encrypt/decrypt the `.secret.env.enc` file, just provide a comma separated string of every team member's **public** key to the creation rule which it applies.

```yaml
creation_rules:
  - path_regex: \.secret\.env\.enc$
    age: >-
      age1jdfjhhdce2wqmgpngc5cx04ljwtnlnn9033p6gkfvmm9fv5ydsastnk8p9,
      age1ngyzntk0nzytotc3mi0xmwyxltkzodetogjhogiyyme0otfjcgnzytotc3,
      age1njm4ntjhmjitotc3mi0xmwyxlwfkymetzwzindg3mgnhnjc4cgnjjiymm3n
```
