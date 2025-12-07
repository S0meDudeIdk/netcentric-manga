import React from 'react';
import { Book, Github, Mail } from 'lucide-react';
import { Link } from 'react-router-dom';

const Footer = () => {
  return (
    <footer className="bg-white dark:bg-[#141118] border-t border-zinc-200 dark:border-zinc-800 transition-colors">
      <div className="container mx-auto px-4 py-12">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-12">
          {/* Brand */}
          <div className="col-span-1 space-y-4">
            <Link to="/" className="flex items-center gap-2 group w-fit">
              <div className="p-2 rounded-lg bg-primary/10 group-hover:bg-primary/20 transition-colors">
                <Book className="w-6 h-6 text-primary" />
              </div>
              <span className="text-xl font-bold text-zinc-900 dark:text-white tracking-tight">MangaHub</span>
            </Link>
            <p className="text-sm text-zinc-500 dark:text-zinc-400 leading-relaxed">
              Your personal manga library and reading tracker. Discover new worlds, track your progress, and join the community.
            </p>
          </div>

          {/* Quick Links */}
          <div>
            <h3 className="font-bold text-zinc-900 dark:text-white mb-6">Quick Links</h3>
            <ul className="space-y-3">
              <li>
                <Link to="/" className="text-sm text-zinc-500 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors">
                  Home
                </Link>
              </li>
              <li>
                <Link to="/browse" className="text-sm text-zinc-500 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors">
                  Browse Manga
                </Link>
              </li>
              <li>
                <Link to="/search" className="text-sm text-zinc-500 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors">
                  Search
                </Link>
              </li>
              <li>
                <Link to="/library" className="text-sm text-zinc-500 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors">
                  My Library
                </Link>
              </li>
            </ul>
          </div>

          {/* About */}
          <div>
            <h3 className="font-bold text-zinc-900 dark:text-white mb-6">Resources</h3>
            <ul className="space-y-3">
              <li>
                <a href="#" className="text-sm text-zinc-500 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors">
                  About Us
                </a>
              </li>
              <li>
                <a href="#" className="text-sm text-zinc-500 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors">
                  Features
                </a>
              </li>
              <li>
                <a href="#" className="text-sm text-zinc-500 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors">
                  API Documentation
                </a>
              </li>
              <li>
                <a href="#" className="text-sm text-zinc-500 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors">
                  Terms of Service
                </a>
              </li>
            </ul>
          </div>

          {/* Contact */}
          <div>
            <h3 className="font-bold text-zinc-900 dark:text-white mb-6">Connect</h3>
            <ul className="space-y-3">
              <li>
                <a
                  href="https://github.com/S0meDudeIdk/netcentric-manga"
                  className="flex items-center gap-2 text-sm text-zinc-500 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <Github className="w-4 h-4" />
                  <span>GitHub</span>
                </a>
              </li>
              <li>
                <a
                  href="mailto:contact@mangahub.com"
                  className="flex items-center gap-2 text-sm text-zinc-500 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors"
                >
                  <Mail className="w-4 h-4" />
                  <span>Email Us</span>
                </a>
              </li>
            </ul>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="border-t border-zinc-200 dark:border-zinc-800 mt-12 pt-8 text-center text-sm text-zinc-500 dark:text-zinc-500">
          <p>&copy; {new Date().getFullYear()} MangaHub. All rights reserved.</p>
          <p className="mt-2">
            Built with React, Tailwind CSS, and ❤️ for manga lovers
          </p>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
