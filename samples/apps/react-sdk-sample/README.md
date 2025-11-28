# Thunder React SDK Sample Application

This sample application demonstrates how to integrate Thunder (WSO2's authentication system) into a React application using the `@asgardeo/react` SDK. The app showcases user authentication, token management, and displays decoded JWT tokens.

## Features

- 🔐 User authentication with Thunder
- 👤 Display user profile information (name, username)
- 🎫 View access tokens and decoded JWT components (header, payload, signature)
- 🎨 Modern UI with Oxygen UI components
- 🔄 Token refresh and session management
- 📱 Responsive design

## Prerequisites

Before you begin, ensure you have the following installed:

- [Node.js](https://nodejs.org/) (v18 or higher)
- [pnpm](https://pnpm.io/) package manager
- A running Thunder server instance

## Getting Started

### 1. Clone the Repository

```bash
git clone <repository-url>
cd samples/apps/react-sdk-sample
```

### 2. Install Dependencies

```bash
pnpm install
```

### 3. Configure Thunder Application

Before running the app, you need to create an application in Thunder:

1. Access your Thunder Admin Portal
2. Navigate to **Applications** section
3. Click **Create Application**
4. Fill in the application details:
   - **Application Name**: Your app name (e.g., "React SDK Sample")
   - **Application Type**: Single Page Application (SPA)
5. Configure the following settings:
   - **Authorized Redirect URLs**: Add `http://localhost:5173` (or your dev server URL)
   - **Allowed Origins**: Add `http://localhost:5173`
   - **Access Token**: Configure token expiration and other settings as needed
6. Save the application and copy the **Client ID**

### 4. Configure Environment Variables

Create a `.env` file in the root of the project by copying the example file:

```bash
cp .env.example .env
```

Edit the `.env` file and add your configuration:

```env
## General App Configuration

# App client ID (from Thunder application)
VITE_REACT_APP_CLIENT_ID=your_client_id_here

# The base URL for the Thunder server
# E.g., https://localhost:8090 or https://your-thunder-domain.com
VITE_THUNDER_BASE_URL=https://localhost:8090
```

### 5. Run the Application

Start the development server:

```bash
pnpm dev
```

The application will be available at [http://localhost:5173](http://localhost:5173)

## How It Works

### Authentication Flow

1. **AsgardeoProvider Setup** ([main.tsx](src/main.tsx))
   - Wraps the entire app with `AsgardeoProvider`
   - Configures Thunder base URL and client ID
   - Sets platform to "AsgardeoV2"

2. **Conditional Rendering** ([App.tsx](src/App.tsx))
   - Uses `SignedIn` component to show content for authenticated users
   - Uses `SignedOut` component to show sign-in button for unauthenticated users

3. **Token Management** ([HomePage.tsx](src/pages/HomePage.tsx))
   - Retrieves access token using `getAccessToken()` hook
   - Decodes JWT token to display header, payload, and signature
   - Extracts user information (given_name, family_name, username)

### Key Components

#### AsgardeoProvider Configuration

```tsx
<AsgardeoProvider
  baseUrl={import.meta.env.VITE_THUNDER_BASE_URL}
  clientId={import.meta.env.VITE_REACT_APP_CLIENT_ID}
  platform="AsgardeoV2"
>
  <App />
</AsgardeoProvider>
```

#### Using Authentication Hooks

```tsx
import { useAsgardeo } from "@asgardeo/react";

const { getAccessToken, signIn } = useAsgardeo();

// Get access token
const accessToken = await getAccessToken();

// Trigger sign-in
await signIn();
```

#### Conditional Content Rendering

```tsx
import { SignedIn, SignedOut, SignInButton } from "@asgardeo/react";

<SignedIn>
  {/* Content for authenticated users */}
  <HomePage />
</SignedIn>

<SignedOut>
  {/* Content for non-authenticated users */}
  <SignInButton />
</SignedOut>
```

## Available Scripts

- `pnpm dev` - Start development server
- `pnpm build` - Build for production
- `pnpm preview` - Preview production build
- `pnpm lint` - Run ESLint

## Troubleshooting

### Common Issues

**Issue**: "Failed to fetch token"
- **Solution**: Ensure Thunder server is running and accessible at the configured base URL
- Check that the client ID is correct
- Verify redirect URLs are properly configured in Thunder

**Issue**: "Invalid client" error
- **Solution**: Double-check the `VITE_REACT_APP_CLIENT_ID` in your `.env` file
- Ensure the application exists in Thunder and is enabled

**Issue**: CORS errors
- **Solution**: Add your application URL to the "Allowed Origins" in Thunder `deployment.yaml`
```
cors:
  allowed_origins:
    ...
    - "http://localhost:5173"
```

