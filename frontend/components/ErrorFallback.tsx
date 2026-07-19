"use client";

import { useState } from 'react';

export default function ErrorFallback() {
  const [showDetails, setShowDetails] = useState(false);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-6">
      <div className="text-center space-y-6">
        <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-red-100 text-red-600">
          <span className="text-2xl">⚠️</span>
        </div>
        <h1 className="text-2xl font-bold text-gray-900">
          Something went wrong.
        </h1>
        <p className="text-gray-600">
          We&apos;ve been notified about this issue and will take a look at it shortly.
        </p>
        <button
          onClick={() => setShowDetails(!showDetails)}
          className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          {showDetails ? 'Hide details' : 'Show details'}
        </button>
        {showDetails && (
          <div className="mt-4 text-left text-sm bg-gray-50 rounded p-4">
            <p className="font-medium">Error details:</p>
            <pre className="mt-2 bg-white p-4 rounded overflow-auto">
              {/* In a real app, you would show the error details here */}
              Error details would be shown in a production environment.
            </pre>
          </div>
        )}
      </div>
    </div>
  );
}
