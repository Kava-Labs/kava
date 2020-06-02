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
            path: "https://docs.kava.io/Tools/oracle.html"
          },
          {
            title: "Auction Bot",
            path: "https://docs.kava.io/Tools/auction.html",
            static: true
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
            path: "https://godoc.org/github.com/Kava-Labs/kava"
          }
        ]
      },
      {
        title: "Community",
        children: [
          {
            title: "Discord",
            path: "https://discord.com/channels/704389840614981673/704389841051320362"
          },
          {
            title: "Telegram",
            path: "https://t.me/kavalabs"
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
