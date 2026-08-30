# MI Bill Tracker

A Go + React application that tracks Michigan Legislature bills scheduled for voting and allows users to contact their representatives via templatized emails.

## Features

- **Bill Tracking**: View Michigan House and Senate bills scheduled for floor or committee votes
- **Committee Meetings**: See upcoming committee meetings with bill agendas
- **Representative Lookup**: Find your Michigan House and Senate representatives by ZIP code
- **Email Templates**: Pre-written templates for supporting, opposing, or requesting information about bills
- **One-Click Emails**: Generate mailto links to contact multiple representatives at once

## Technology Stack

- **Backend**: Go 1.25 with standard library HTTP server
- **Frontend**: React 18, TypeScript, Vite, Tailwind CSS, Radix UI
- **Data**: Michigan Legislature website (with sample data for demonstration)

## Quick Start

### Prerequisites

- Go 1.25+ (for local development)
- Node.js 18+ (for local development)
- Docker & Docker Compose (for containerized deployment)

### Docker Deployment (Recommended)

1. Copy the environment template:
```bash
cp .env.example .env
```

2. Edit `.env` to configure your settings (optional - defaults work out of the box)

3. Build and start with Docker Compose:
```bash
docker compose up -d
```

4. Open your browser to `http://localhost:8080`

To stop:
```bash
docker compose down
```

### Running Locally (Development)

1. Build the frontend:
```bash
cd web
npm install
npm run build
cd ..
```

2. Build and run the Go server:
```bash
go build ./cmd/server
./server
```

3. Open your browser to `http://localhost:8080`

### Development Mode

Run the frontend dev server:
```bash
cd web
npm run dev
```

Run the Go backend (in another terminal):
```bash
go run ./cmd/server
```

The frontend will proxy API requests to `http://localhost:8080`.

## API Endpoints

- `GET /api/bills` - List all bills
- `GET /api/bills/scheduled` - List bills with scheduled votes
- `GET /api/bills/:id` - Get bill details
- `GET /api/meetings` - List committee meetings
- `POST /api/representatives` - Find representatives by location
- `GET /api/representatives/search?q=` - Search representatives by name
- `GET /api/email/templates` - List email templates
- `POST /api/email/compose` - Generate mailto links
- `POST /api/email/preview` - Preview email content

## Project Structure

```
mibilltracker/
├── cmd/server/           # Go application entry point
├── internal/
│   ├── api/             # HTTP handlers
│   ├── bills/           # Bill models and service
│   ├── representatives/ # Representative lookup service
│   └── email/           # Email templating service
└── web/                 # React frontend
    ├── src/
    │   ├── api/        # API client
    │   ├── components/ # React components
    │   ├── pages/      # Page components
    │   └── types/      # TypeScript types
    └── dist/           # Production build
```

## Data Sources

The application includes sample bill and representative data for demonstration. To use real Michigan Legislature data:

1. Sign up for a free [LegiScan API](https://legiscan.com/legiscan) key
2. Set the `LEGISCAN_API_KEY` environment variable
3. Update `internal/bills/service.go` to use `ScrapeLegiScanBills()`

For production representative lookup, integrate with:
- Michigan District maps API
- Google Civic Information API
- Or host district boundary data locally

## Configuration

All configuration is managed through the `.env` file. Copy `.env.example` to `.env` and customize:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `LEGISCAN_API_KEY` | LegiScan API key for real Michigan bill data | (empty) |
| `FETCH_INTERVAL_MINUTES` | How often to refresh bill data | `15` |

## Email Templates

The application includes three default templates:
- **Support Bill**: Express support for legislation
- **Oppose Bill**: Express opposition to legislation
- **Request Information**: Ask questions about pending legislation

Templates use Go's `text/template` syntax with variables like `{{.BillNumber}}`, `{{.RepresentativeName}}`, etc.

## License

MIT
