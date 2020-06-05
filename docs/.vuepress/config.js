module.exports = {
  theme: "cosmos",
  title: "Kava Documentation",
  themeConfig: {
    logo: {
      src: "/kava-logo.svg",
    },
    custom: true,
    autoSidebar: true,
    sidebar: [
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
  }
};
