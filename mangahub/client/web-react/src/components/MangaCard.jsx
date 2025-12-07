import React from 'react';
import { motion } from 'framer-motion';
import { Book, Star, Calendar } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const MangaCard = ({ manga, index = 0 }) => {
  const navigate = useNavigate();

  const handleClick = () => {
    navigate(`/manga/${manga.id}`);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay: index * 0.05 }}
      whileHover={{ y: -8, scale: 1.02 }}
      className="group relative bg-white dark:bg-[#211c27] rounded-xl overflow-hidden cursor-pointer shadow-sm hover:shadow-xl hover:shadow-primary/10 border border-zinc-100 dark:border-zinc-800 transition-all duration-300"
      onClick={handleClick}
    >
      {/* Cover Image Container */}
      <div className="relative aspect-[2/3] overflow-hidden bg-zinc-100 dark:bg-zinc-800">
        {manga.cover_url ? (
          <img
            src={manga.cover_url}
            alt={manga.title}
            className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
            loading="lazy"
            onError={(e) => {
              e.target.src = 'https://via.placeholder.com/300x400?text=No+Cover';
            }}
          />
        ) : (
          <div className="w-full h-full flex flex-col items-center justify-center bg-zinc-100 dark:bg-zinc-800 text-zinc-300 dark:text-zinc-600">
            <Book className="w-12 h-12 mb-2" />
            <span className="text-sm font-medium">No Cover</span>
          </div>
        )}

        {/* Overlay Gradient */}
        <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

        {/* Rating Badge */}
        {manga.rating && (
          <div className="absolute top-2 right-2 bg-black/60 backdrop-blur-md text-white px-2 py-1 rounded-lg flex items-center gap-1 text-xs font-bold shadow-lg border border-white/10">
            <Star className="w-3 h-3 fill-yellow-400 text-yellow-400" />
            <span>{manga.rating.toFixed(1)}</span>
          </div>
        )}

        {/* Status Badge (shows on hover) */}
        {manga.status && (
          <div className="absolute bottom-2 left-2 translate-y-4 opacity-0 group-hover:translate-y-0 group-hover:opacity-100 transition-all duration-300">
            <span
              className={`inline-block px-2 py-1 text-[10px] font-bold uppercase tracking-wider rounded-md border ${manga.status === 'completed'
                  ? 'bg-green-500/80 text-white border-green-400/50'
                  : manga.status === 'ongoing'
                    ? 'bg-blue-500/80 text-white border-blue-400/50'
                    : 'bg-zinc-500/80 text-white border-zinc-400/50'
                } backdrop-blur-md shadow-lg`}
            >
              {manga.status}
            </span>
          </div>
        )}
      </div>

      {/* Content */}
      <div className="p-4">
        <h3 className="font-bold text-zinc-900 dark:text-white mb-1 truncate leading-tight group-hover:text-primary transition-colors" title={manga.title}>
          {manga.title}
        </h3>

        {/* Genres */}
        <div className="flex flex-wrap gap-1 mb-2 h-5 overflow-hidden">
          {manga.genres?.slice(0, 2).map((genre, i) => (
            <span key={i} className="text-[10px] text-zinc-500 dark:text-zinc-400 font-medium">
              {genre}{i < Math.min(manga.genres.length, 2) - 1 ? " • " : ""}
            </span>
          ))}
        </div>

        {/* Meta Info */}
        <div className="flex items-center justify-between mt-3 pt-3 border-t border-zinc-100 dark:border-zinc-800">
          <div className="flex items-center gap-1.5 text-xs text-zinc-500 dark:text-zinc-400">
            <Book className="w-3.5 h-3.5" />
            <span>{manga.total_chapters || "?"} ch</span>
          </div>

          {manga.publication_year && (
            <div className="flex items-center gap-1.5 text-xs text-zinc-500 dark:text-zinc-400">
              <Calendar className="w-3.5 h-3.5" />
              <span>{manga.publication_year}</span>
            </div>
          )}
        </div>
      </div>
    </motion.div>
  );
};

export default MangaCard;
