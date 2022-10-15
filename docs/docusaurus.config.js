// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require("prism-react-renderer/themes/github");
const darkCodeTheme = require("prism-react-renderer/themes/dracula");

// const lastVersion = "v0.47";
const lastVersion = "current";

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "MediBloc Panacea",
  tagline:
    "MediBloc Panacea",
  url: "https://docs.gopanacea.org",
  baseUrl: "/",
  onBrokenLinks: "warn",
  onBrokenMarkdownLinks: "warn",
  favicon: "img/favicon.png",
  trailingSlash: false,

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: "medibloc",
  projectName: "panacea-core",

  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "classic",
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          routeBasePath: "/",
          lastVersion: lastVersion,
          versions: {
            current: {
              path: "master",
              // banner: "unreleased",
            },
            // "v0.47": {
            //   label: "v0.47",
            //   path: "v0.47",
            //   banner: "none",
            // },
          },
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      image: "img/banner.png",
      docs: {
        sidebar: {
          autoCollapseCategories: true,
        },
      },
      navbar: {
        title: "MediBloc Panacea",
        hideOnScroll: false,
        logo: {
          alt: "MediBloc Panacea Logo",
          src: "img/logo.png",
          href: "https://docs.gopanacea.org",
          target: "_self",
        },
        items: [
          {
            href: "https://github.com/medibloc/panacea-core",
            html: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" class="github-icon">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M12 0.300049C5.4 0.300049 0 5.70005 0 12.3001C0 17.6001 3.4 22.1001 8.2 23.7001C8.8 23.8001 9 23.4001 9 23.1001C9 22.8001 9 22.1001 9 21.1001C5.7 21.8001 5 19.5001 5 19.5001C4.5 18.1001 3.7 17.7001 3.7 17.7001C2.5 17.0001 3.7 17.0001 3.7 17.0001C4.9 17.1001 5.5 18.2001 5.5 18.2001C6.6 20.0001 8.3 19.5001 9 19.2001C9.1 18.4001 9.4 17.9001 9.8 17.6001C7.1 17.3001 4.3 16.3001 4.3 11.7001C4.3 10.4001 4.8 9.30005 5.5 8.50005C5.5 8.10005 5 6.90005 5.7 5.30005C5.7 5.30005 6.7 5.00005 9 6.50005C10 6.20005 11 6.10005 12 6.10005C13 6.10005 14 6.20005 15 6.50005C17.3 4.90005 18.3 5.30005 18.3 5.30005C19 7.00005 18.5 8.20005 18.4 8.50005C19.2 9.30005 19.6 10.4001 19.6 11.7001C19.6 16.3001 16.8 17.3001 14.1 17.6001C14.5 18.0001 14.9 18.7001 14.9 19.8001C14.9 21.4001 14.9 22.7001 14.9 23.1001C14.9 23.4001 15.1 23.8001 15.7 23.7001C20.5 22.1001 23.9 17.6001 23.9 12.3001C24 5.70005 18.6 0.300049 12 0.300049Z" fill="currentColor"/>
            </svg>
            `,
            position: "right",
          },
          {
            type: "docsVersionDropdown",
            position: "left",
            dropdownActiveClassDisabled: true,
            // versions not yet migrated to docusaurus
            dropdownItemsAfter: [
              {
                href: "https://docs.gopanacea.org/v2.0/",
                label: "v2.0",
                target: "_self",
              }
            ],
          },
        ],
      },
      footer: {
        links: [
          {
            items: [
              {
                html: `<a href="https://docs.gopanacea.org"><img src="/img/logo-bw.png" alt="MediBloc Panacea Logo"></a>`,
              },
            ],
          },
          {
            title: "Community",
            items: [
              {
                label: "Blog",
                href: "https://blog.medibloc.org/",
              },
              {
                label: "Twitter",
                href: "https://twitter.com/_medibloc",
              },
              {
                label: "Discord",
                href: "https://discord.gg/gHVnczcQ3E",
              }
            ],
          }
        ],
        copyright: `<p>© 2022 MediBloc Limited. All right reserved. Suite 23, Portland House, Glacis Road, GX11 1AA, Gibraltar</p>`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
        additionalLanguages: ["protobuf", "go-module"], // https://prismjs.com/#supported-languages
      },
      // algolia: {
      //   appId: "BH4D9OD16A",
      //   apiKey: "ac317234e6a42074175369b2f42e9754",
      //   indexName: "cosmos-sdk",
      //   contextualSearch: false,
      // },
    }),
  themes: ["@you54f/theme-github-codeblock"],
  plugins: [
    async function myPlugin(context, options) {
      return {
        name: "docusaurus-tailwindcss",
        configurePostCss(postcssOptions) {
          postcssOptions.plugins.push(require("postcss-import"));
          postcssOptions.plugins.push(require("tailwindcss/nesting"));
          postcssOptions.plugins.push(require("tailwindcss"));
          postcssOptions.plugins.push(require("autoprefixer"));
          return postcssOptions;
        },
      };
    },
  ],
};

module.exports = config;
