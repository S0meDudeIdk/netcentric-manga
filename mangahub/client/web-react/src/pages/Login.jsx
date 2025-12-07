import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Eye, EyeOff, AlertCircle } from 'lucide-react';
import authService from '../services/authService';

const Login = () => {
  const [formData, setFormData] = useState({
    email: '',
    password: ''
  });
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    });
    setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await authService.login({
        email: formData.email,
        password: formData.password
      });
      navigate('/library');
    } catch (err) {
      setError(err.error || err.message || 'Invalid email or password');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="relative flex min-h-screen w-full flex-col group/design-root">
      <div className="flex-grow w-full">
        <div className="min-h-screen flex flex-row">
          {/* Left Panel: Illustration */}
          <div className="hidden lg:flex w-1/2 items-center justify-center p-12 bg-background-light dark:bg-[#141118] illustration-bg">
            <div className="w-full max-w-md">
              <img
                className="w-full h-auto"
                alt="Stylized illustration of manga characters"
                src="/assets/auth_hero.png"
              />
            </div>
          </div>

          {/* Right Panel: Form */}
          <div className="w-full lg:w-1/2 flex items-center justify-center p-6 sm:p-12 bg-background-light dark:bg-background-dark">
            <div className="flex flex-col w-full max-w-md">
              {/* Logo */}
              <div className="flex items-center gap-3 mb-8">
                <svg fill="none" height="32" viewBox="0 0 32 32" width="32" xmlns="http://www.w3.org/2000/svg">
                  <path className="stroke-zinc-900 dark:stroke-white" d="M16 31.5C24.5604 31.5 31.5 24.5604 31.5 16C31.5 7.43959 24.5604 0.5 16 0.5C7.43959 0.5 0.5 7.43959 0.5 16C0.5 24.5604 7.43959 31.5 16 31.5Z" strokeWidth="1"></path>
                  <path className="stroke-zinc-900 dark:stroke-white" d="M16 31.5C24.5604 31.5 31.5 24.5604 31.5 16C31.5 7.43959 24.5604 0.5 16 0.5C7.43959 0.5 0.5 7.43959 0.5 16C0.5 24.5604 7.43959 31.5 16 31.5Z" strokeWidth="1"></path>
                  <path className="stroke-primary fill-primary/20" d="M11 11.25H21V20.75L16 26L11 20.75V11.25Z" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"></path>
                </svg>
                <h2 className="text-xl font-bold text-zinc-900 dark:text-white">MangaHub</h2>
              </div>

              {/* Page Heading */}
              <div className="flex min-w-72 flex-col gap-2 mb-8">
                <p className="text-zinc-900 dark:text-white text-4xl font-black leading-tight tracking-[-0.033em]">Welcome Back!</p>
                <p className="text-zinc-500 dark:text-[#ab9db9] text-base font-normal leading-normal">Log in to continue to MangaHub.</p>
              </div>

              {/* Error Message */}
              {error && (
                <div className="mb-6 p-4 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-900/50 flex items-start gap-3">
                  <AlertCircle className="w-5 h-5 text-red-600 dark:text-red-400 mt-0.5" />
                  <p className="text-sm text-red-600 dark:text-red-400 font-medium">{error}</p>
                </div>
              )}

              <form onSubmit={handleSubmit} className="flex flex-col gap-4">
                <label className="flex flex-col min-w-40 flex-1">
                  <p className="text-zinc-900 dark:text-white text-base font-medium leading-normal pb-2">Username or Email</p>
                  <input
                    name="email"
                    type="text" // Changed to text to match HTML "Username or Email" placeholder intent, though name is email
                    value={formData.email}
                    onChange={handleChange}
                    className="form-input flex w-full min-w-0 flex-1 resize-none overflow-hidden rounded-lg text-zinc-900 dark:text-white focus:outline-0 focus:ring-2 focus:ring-primary/50 border border-zinc-300 dark:border-[#473b54] bg-white dark:bg-[#211c27] focus:border-primary dark:focus:border-primary h-14 placeholder:text-zinc-400 dark:placeholder:text-[#ab9db9] p-[15px] text-base font-normal leading-normal"
                    placeholder="Enter your username or email"
                  />
                </label>

                <div className="flex flex-col min-w-40 flex-1">
                  <div className="flex items-center justify-between pb-2">
                    <label className="text-zinc-900 dark:text-white text-base font-medium leading-normal" htmlFor="password">Password</label>
                    <a className="text-sm font-semibold text-primary hover:underline" href="#">Forgot Password?</a>
                  </div>
                  <div className="flex w-full flex-1 items-stretch rounded-lg group">
                    <input
                      id="password"
                      name="password"
                      type={showPassword ? "text" : "password"}
                      value={formData.password}
                      onChange={handleChange}
                      className="form-input flex w-full min-w-0 flex-1 resize-none overflow-hidden rounded-lg text-zinc-900 dark:text-white focus:outline-0 focus:ring-2 focus:ring-primary/50 border border-zinc-300 dark:border-[#473b54] bg-white dark:bg-[#211c27] focus:border-primary dark:focus:border-primary h-14 placeholder:text-zinc-400 dark:placeholder:text-[#ab9db9] p-[15px] rounded-r-none border-r-0 pr-2 text-base font-normal leading-normal"
                      placeholder="Enter your password"
                    />
                    <div
                      onClick={() => setShowPassword(!showPassword)}
                      className="text-zinc-400 dark:text-[#ab9db9] flex border border-zinc-300 dark:border-[#473b54] bg-white dark:bg-[#211c27] items-center justify-center px-4 rounded-r-lg border-l-0 group-focus-within:border-primary group-focus-within:ring-2 group-focus-within:ring-primary/50 cursor-pointer"
                    >
                      {showPassword ? <EyeOff className="w-6 h-6" /> : <Eye className="w-6 h-6" />}
                    </div>
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="flex min-w-[84px] cursor-pointer items-center justify-center overflow-hidden rounded-lg h-14 px-5 flex-1 bg-primary text-white text-base font-bold leading-normal tracking-[0.015em] hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary/50 dark:focus:ring-offset-background-dark mt-4 disabled:opacity-70 disabled:cursor-not-allowed"
                >
                  <span className="truncate">{loading ? 'Logging In...' : 'Log In'}</span>
                </button>
              </form>

              <div className="mt-8 text-center text-sm text-zinc-600 dark:text-zinc-400">
                Don't have an account?{' '}
                <Link className="font-semibold text-primary hover:underline" to="/register">Sign Up</Link>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;
