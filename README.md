# Finance Tracker

Finance Tracker is a backend application designed to help users manage their personal finances. It allows for tracking income, expenses, debts, and savings goals, and provides summaries and reports. It also includes features for financial advice via an AI service and notifications for upcoming due dates.

## Features

*   **Income Management**: Record and categorize income.
*   **Expense Tracking**: Log and categorize expenses.
*   **Debt Management**: Keep track of debts, due dates, and statuses.
*   **Savings Goals**: Set and monitor progress towards savings goals.
*   **Financial Summaries**: Generate weekly, monthly, and yearly financial summaries (total income, total expenses, net balance).
*   **Reporting**:
    *   Generate transaction reports in CSV format.
    *   Generate PDF reports summarizing transactions and financial standing.
*   **AI-Powered Advice**: Get financial advice based on summaries using the OpenRouter API.
*   **Notifications**: Receive reminders for upcoming debt due dates and approaching savings goal target dates.

## Tech Stack

*   **Backend**: Go
*   **Database**: SQL (interfaced via GORM ORM)
*   **AI Integration**: OpenRouter API (specifically using `mistralai/mistral-7b-instruct:free` as an example model)
*   **PDF Generation**: `jung-kurt/gofpdf`

## Project Structure (Internal Services)

The `internal/services` directory contains the core logic for the application:

*   `ai_advice_service.go`: Handles interaction with the OpenRouter API to provide financial advice.
*   `debt_service.go`: Manages CRUD operations and logic for debts.
*   `expense_service.go`: Manages CRUD operations and logic for expenses.
*   `income_service.go`: Manages CRUD operations and logic for income.
*   `notification_service.go`: Handles scheduled checks and notifications for debts and savings goals.
*   `report_service.go`: Generates CSV and PDF financial reports.
*   `savings_service.go`: Manages CRUD operations and logic for savings goals.
*   `summary_service.go`: Calculates and stores/retrieves financial summaries.
*   `summary_service_test.go`: Contains unit tests for the summary service, particularly for period calculations.

## Setup and Installation

1.  **Prerequisites**:
    *   Go (version X.Y.Z or higher recommended)
    *   A running SQL database compatible with GORM (e.g., PostgreSQL, MySQL, SQLite).
    *   Set up environment variables (see `.env.example` if provided, or below).

2.  **Clone the repository**:
    ```bash
    git clone <your-repository-url>
    cd finance-tracker
    ```

3.  **Install dependencies**:
    ```bash
    go mod tidy
    ```

4.  **Environment Variables**:
    Create a `.env` file in the root of the project (or configure your environment) with the following variables:
    *   `DB_HOST`: Database host
    *   `DB_PORT`: Database port
    *   `DB_USER`: Database username
    *   `DB_PASSWORD`: Database password
    *   `DB_NAME`: Database name
    *   `DB_SSLMODE`: (e.g., `disable`, `require`)
    *   `OPENROUTER_API_KEY`: Your API key for OpenRouter.ai (Optional, for AI advice feature. Can be set to `YOUR_DUMMY_OPENROUTER_API_KEY_FOR_TESTING` for basic testing without live API calls).

5.  **Database Migrations**:
    Ensure the database schema is set up. The application uses GORM, which can handle migrations. You might need to run a migration command if provided, or GORM might auto-migrate based on your models upon the first run (depending on configuration in `internal/database/database.go`).

6.  **Run the application**:
    ```bash
    go run cmd/server/main.go  # Or your main entry point
    ```

## API Endpoints

(To be documented - list your API endpoints here if this is an API server)

*   `GET /incomes`: Retrieves a list of incomes.
*   `POST /incomes`: Creates a new income entry.
*   ... and so on for expenses, debts, savings, reports, summaries.

## Usage

(Describe how a user or another service would interact with your application. If it's a CLI, provide CLI commands. If it's an API, provide example requests.)

### Generating Summaries

The system can generate weekly, monthly, or yearly financial summaries. These are typically created or fetched on demand.

### AI Financial Advice

If the `OPENROUTER_API_KEY` is configured, the application can provide financial advice based on the generated summaries.

### Notifications

The notification service runs periodically (configuration not shown in provided files, but typically a cron job or scheduled task) to check for:
*   Debts due within the next 7 days.
*   Savings goals approaching their target date within the next 7 days where the current amount is less than the goal amount.
Reminders are logged by the service.

## Testing

Unit tests are included for some services. To run tests:

```bash
go test ./...
