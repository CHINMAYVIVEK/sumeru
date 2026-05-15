# Web Icons

Sumeru uses an SVG sprite system for menu icons. This approach is highly performant as it requires only a single HTTP request, is easily cached by the browser, and avoids the overhead of a heavy Javascript icon library.

## How it works

When you define a `<menuitem>` in an XML file, the `web_icon` attribute determines which icon is displayed in the sidebar or top bar:

```xml
<menuitem id="menu_my_app" name="My App" web_icon="package"/>
```

The system maps the `web_icon` value (e.g., `package`) to a symbol ID in the global SVG sprite file:
`core/engine/assets/img/menu-icons.svg#i-package`

## Available Icons

The following icons are currently available in the core sprite (based on **Lucide**):

### General & Navigation
| Icon Key | Description |
| :--- | :--- |
| `home` | Home / Dashboard |
| `apps` | App switcher / Grid |
| `search` | Search operations |
| `bell` | Notifications |
| `chevron-down` | Dropdown indicator |
| `chevron-right` | Sidebar / Nested indicator |
| `more-horizontal` | Actions menu |
| `refresh-cw` | Reload / Sync |
| `external-link` | Outbound links |
| `eye` | View / Show |
| `eye-off` | Hide / Private |

### Finance & Accounting
| Icon Key | Description |
| :--- | :--- |
| `banknote` | Cash / Payments |
| `coins` | Currency / Change |
| `receipt` | Invoicing / Expenses |
| `wallet` | Balance / Wallets |
| `trending-up` | Growth / Profit |
| `trending-down` | Loss / Decrease |
| `pie-chart` | Allocation / Analysis |
| `bar-chart-3` | Reporting / Statistics |
| `percent` | Taxes / Discounts |
| `credit-card` | Cards / Billing |

### Sales & CRM
| Icon Key | Description |
| :--- | :--- |
| `cart` | Shopping cart |
| `contact` | Leads / Contacts |
| `phone` | Calls / Communication |
| `message-square` | Chat / Feedback |
| `target` | Goals / Leads |
| `gift` | Rewards / Promotions |
| `tag` | Pricing / Labels |
| `ticket` | Support / Vouchers |

### HR & People
| Icon Key | Description |
| :--- | :--- |
| `users` | Groups / Teams |
| `user` | Individual account |
| `user-plus` | Onboarding / Recruitment |
| `user-minus` | Offboarding |
| `user-check` | Validation / Attendance |
| `user-cog` | Profile Settings |
| `briefcase` | Positions / Jobs |
| `graduation-cap` | Training / Skills |
| `heart` | Benefits / Health |

### Operations & Manufacturing
| Icon Key | Description |
| :--- | :--- |
| `inventory` | Stock / Assets |
| `package` | Products / Items |
| `truck` | Shipping / Logistics |
| `factory` | Production / Plants |
| `settings-2` | Adjustments |
| `wrench` | Maintenance / Tools |
| `hammer` | Construction / Work |
| `gauge` | Performance / Metrics |
| `timer` | Tracking / Process |

### Project Management
| Icon Key | Description |
| :--- | :--- |
| `clock` | Time tracking |
| `milestone` | Roadmap / Stages |
| `kanban` | Task boards |
| `clipboard-list` | Audits / Tasks |
| `check-square` | Completion |
| `flag` | Priority / Markers |

### Data & Files
| Icon Key | Description |
| :--- | :--- |
| `folder` | Directories |
| `file-text` | Records / PDF |
| `database` | System / Servers |
| `archive` | History / Vault |
| `inbox` | Communications |
| `send` | Dispatch / Transmit |
| `paperclip` | Attachments |
| `share-2` | Social / Export |
| `download` | Export / Get |
| `upload` | Import / Put |

### System & Security
| Icon Key | Description |
| :--- | :--- |
| `settings` / `cog` | Configuration |
| `shield` | Security / Access |
| `lock` | Locked / Restricted |
| `unlock` | Open / Public |
| `layers` | Framework / Stack |
| `layout` | UI / Blueprints |
| `log` | Event logs |
| `info` | Help / Details |
| `alert-circle` | Warnings |
| `help-circle` | FAQ |

## Adding New Icons

To add a new icon to the system:

1.  **Find the Icon**: Go to [Lucide.dev](https://lucide.dev) and find the icon you need.
2.  **Get the SVG Path**: Copy the SVG paths (`<path>`, `<circle>`, etc.) from the Lucide site.
3.  **Update the Sprite**: Open `sumeru/core/engine/assets/img/menu-icons.svg`.
4.  **Add a `<symbol>`**: Create a new symbol at the end of the `<defs>` section:
    ```xml
    <symbol id="i-your-icon-name" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
      <!-- Paste the paths here -->
    </symbol>
    ```
5.  **Use it**: You can now use `web_icon="your-icon-name"` in any menu definition.

> [!TIP]
> Always use `i-` prefix for the symbol ID in the SVG file, but omit it when referencing the icon in the `web_icon` attribute.
