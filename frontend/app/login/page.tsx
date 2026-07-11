"use client";

import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';

// Define the form schema
const loginSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(1, 'Password is required'),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: '',
      password: '',
    },
  });

  const onSubmit = (data: LoginFormValues) => {
    // In a real app, you would send this data to an authentication endpoint
    // For MVP, we'll use hardcoded credentials from environment variables
    const validUsername = process.env.NEXT_PUBLIC_ADMIN_USER || 'admin';
    const validPassword = process.env.NEXT_PUBLIC_ADMIN_PASS || 'password';

    if (data.username === validUsername && data.password === validPassword) {
      // Set a cookie (non-httpOnly) for authentication
      document.cookie = 'token=authenticated; Path=/; Max-Age=3600';
      // Redirect to dashboard
      router.push('/dashboard');
    } else {
      // In a real app, you would set a form error
      // For now, we'll just alert
      alert('Invalid credentials');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="w-full max-w-md space-y-6">
        <div className="text-center">
          <h1 className="block text-3xl font-bold text-gray-900">
            Story Video Bot
          </h1>
          <p className="mt-2 text-sm text-gray-600">
            Sign in to continue
          </p>
        </div>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label htmlFor="username" className="sr-only">
              Username
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <div className="h-5 w-5 text-gray-400">👤</div>
              </div>
              <input
                id="username"
                type="text"
                name="username"
                {...register('username')}
                required
                className={`block w-full pl-10 pr-3 py-3 pt-2 pb-2 text-sm text-gray-900 bg-white border-0 rounded-md appearence-none focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent sm:text-sm ${
                  errors.username ? 'border-red-500' : ''
                }`}
                placeholder="Username"
              />
              {errors.username && (
                <p className="mt-1 text-sm text-red-600">
                  {errors.username.message}
                </p>
              )}
            </div>
          </div>
          <div>
            <label htmlFor="password" className="sr-only">
              Password
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <div className="h-5 w-5 text-gray-400">🔒</div>
              </div>
              <input
                id="password"
                type="password"
                name="password"
                {...register('password')}
                required
                className={`block w-full pl-10 pr-3 py-3 pt-2 pb-2 text-sm text-gray-900 bg-white border-0 rounded-md appearence-none focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent sm:text-sm ${
                  errors.password ? 'border-red-500' : ''
                }`}
                placeholder="Password"
              />
              {errors.password && (
                <p className="mt-1 text-sm text-red-600">
                  {errors.password.message}
                </p>
              )}
            </div>
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <input
                id="remember-me"
                type="checkbox"
                {...register('rememberMe')}
                className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
              />
              <label htmlFor="remember-me" className="ml-2 block text-sm text-gray-900">
                Remember me
              </label>
            </div>
            <div className="text-sm">
              <a href="#" className="font-medium text-indigo-600 hover:text-indigo-500">
                Forgot password?
              </a>
            </div>
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full justify-center items-center px-4 py-2 text-sm font-medium leading-5 text-white transition-colors duration-150 bg-purple-600 border border-transparent rounded-lg active:bg-purple-600 hover:bg-purple-700 focus:outline-none focus:shadow-outline-purple"
          >
            {isSubmitting ? 'Signing in...' : 'Sign in'}
          </button>
          </form>
        <div className="text-center">
          <p className="text-sm text-gray-500">
            Don't have an account?{' '}
            <a href="#" className="font-medium text-indigo-600 hover:text-indigo-500">
              Sign up
            </a>
          </p>
        </div>
      </div>
    </div>
  );
}