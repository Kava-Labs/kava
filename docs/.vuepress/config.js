module.exports = {
  theme: "cosmos",
  title: "Kava Documentation",
  themeConfig: {
    logo: {
      src: "/kava-logo.svg",
    },
    custom: true,
    autoSidebar: false,
    sidebar: [
      {
        title: "Modules",
        children: [
          {
            title: "CDP",
            path: "./public/favicon.ico"
          },
          {
            title: "Auction",
            path: "https://example.org/"
          },
          {
            title: "BEP3",
            path: "https://example.org/"
          },
          {
            title: "Pricefeed",
            path: "https://example.org/"
          },
          {
            title: "Committee",
            path: "https://example.org/"
          },
          {
            title: "Incentive",
            path: "https://example.org/"
          },
          {
            title: "Kavadist",
            path: "https://example.org/"
          },
          {
            title: "Validator Vesting",
            path: "https://example.org/"
          }
        ]
      },
      {
        title: "Kava Tools",
        children: [
          {
            title: "Chainlink Price Oracle",
            // path: "https://example.org/"
            path: "./dist/kava-tools/oracle.html",
            static: true
          },
          {
            title: "Auction Bot",
            path: "https://example.org/"
          }
        ]
      },
      {
        title: "Building on Kava",
        children: [
          {
            title: "JavaScript SDK",
            path: "https://example.org/"
          }
        ]        
      }
    ]
  }
};
