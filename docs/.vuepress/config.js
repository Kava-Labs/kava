module.exports = {
  theme: "cosmos",
  title: "Kava Documentation",
  themeConfig: {
    logo: {
      src: "/kava-logo.svg",
    },
    custom: true,
    sidebar: {
      auto: false,
      nav: [
        {
          title: "Reference",
          children: [
            {
              title: "Modules",
              path: "/Modules",
              directory: true
            },
          ]
        },
        {
          title: "Kava Tools",
          children: [
            {
              title: "Chainlink Price Oracle",
              path: "/tools/oracle.html"
            },
            {
              title: "Auction Bot",
              path: "/tools/auction.html",
            }
          ]
        },
        {
          title: "Building on Kava",
          children: [
            {
              title: "JavaScript SDK",
              path: "/building/javascript-sdk.html"
            },
            {
              title: "Migration Guide: kava-3",
              path: "/building/kava-3-migration-guide.html"
            }
          ]
        },
        {
          title: "Resources",
          children: [
            {
              title: "REST API Spec",
              path: "https://rpc.kava.io/"
            },
            {
              title: "Protocol Reference",
              path: "https://pkg.go.dev/github.com/kava-labs/kava?tab=subdirectories"
            }
          ]
        },
        {
          title: "Community",
          children: [
            {
              title: "Twitter",
              path: "https://twitter.com/kava_labs"
            },
            {
              title: "Telegram",
              path: "https://t.me/kavalabs"
            },
            {
              title: "Discord",
              path: "https://discord.com/channels/704389840614981673/704389841051320362"
            },
            {
              title: "GitHub",
              path: "https://github.com/Kava-Labs/kava"
            }
          ]
        }
      ]
    },
    footer: {
      logo: "/kava-logo.svg", // recommended ratio: 120 W x 48 H
      textLink: {
        text: "kava.io",
        url: "https://www.kava.io"
      },
      services: [
        {
          service: "twitter",
          url: "https://twitter.com/kava_labs"
        },
        {
          service: "medium",
          url: "https://medium.com/kava-labs"
        },
        {
          service: "telegram",
          url: "https://t.me/kavalabs"
        },
        {
          service: "discord",
          url: "https://discord.com/invite/kQzh3Uv"
        },
        {
          service: "github",
          url: "https://github.com/kava-labs"
        }
      ],
      smallprint:
        `© ${new Date().getFullYear()} Kava Labs.`,
      links: [
        {
          title: "Documentation",
          children: [
            {
              title: "Cosmos SDK",
              url: "https://cosmos.network/docs"
            }
          ]
        },
        {
          title: "Community",
          children: [
            {
              title: "Blog",
              url: "https://medium.com/kava-labs"
            },
            {
              title: "Chat",
              url: "https://discord.com/invite/kQzh3Uv"
            }
          ]
        },
        {
          title: "Contributing",
          children: [
            {
              title: "Contributing to the docs",
              url:
                "https://github.com/Kava-Labs/kava/tree/master/docs"
            },
            {
              title: "Source code on GitHub",
              url: "https://github.com/Kava-Labs/kava"
            }
          ]
        }
      ]
    }
  }
};
