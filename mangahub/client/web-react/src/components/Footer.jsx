import React from 'react';
import { Book, Github, Mail } from 'lucide-react';
import { Link } from 'react-router-dom';

const Footer = () => {
  return (
    <footer className="bg-gray-900 text-gray-300 mt-auto">
      <div className="container mx-auto px-4 py-8">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {/* Brand */}
          <div className="col-span-1">
            <div className="flex items-center gap-2 text-white mb-4">
              <Book className="w-6 h-6" />
              <span className="text-xl font-bold">MangaHub</span>
            </div>
            <p className="text-sm text-gray-400">
              Your personal manga library and reading tracker.
            </p>
          </div>

          {/* Quick Links */}
          <div>
            <h3 className="text-white font-semibold mb-4">Quick Links</h3>
            <ul className="space-y-2">
              <li>
                <Link to="/" className="text-sm hover:text-white transition">
                  Home
                </Link>
              </li>
              <li>
                <Link to="/browse" className="text-sm hover:text-white transition">
                  Browse Manga
                </Link>
              </li>
              <li>
                <Link to="/search" className="text-sm hover:text-white transition">
                  Search
                </Link>
              </li>
              <li>
                <Link to="/library" className="text-sm hover:text-white transition">
                  My Library
                </Link>
              </li>
            </ul>
          </div>

          {/* About */}
          <div>
            <h3 className="text-white font-semibold mb-4">About</h3>
            <ul className="space-y-2">
              <li>
                <a href="#" className="text-sm hover:text-white transition">
                  About Us
                </a>
              </li>
              <li>
                <a href="#" className="text-sm hover:text-white transition">
                  Features
                </a>
              </li>
              <li>
                <a href="#" className="text-sm hover:text-white transition">
                  API Documentation
                </a>
              </li>
              <li>
                <a href="#" className="text-sm hover:text-white transition">
                  Terms of Service
                </a>
              </li>
            </ul>
          </div>

          {/* Contact */}
          <div>
            <h3 className="text-white font-semibold mb-4">Contact</h3>
            <ul className="space-y-2">
              <li>
                <a
                  href="https://github.com/S0meDudeIdk/netcentric-manga"
                  className="flex items-center gap-2 text-sm hover:text-white transition"
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
                  className="flex items-center gap-2 text-sm hover:text-white transition"
                >
                  <Mail className="w-4 h-4" />
                  <span>Email Us</span>
                </a>
              </li>
            </ul>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="border-t border-gray-800 mt-8 pt-6 text-center text-sm text-gray-400">
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
