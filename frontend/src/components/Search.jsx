import { useState, useEffect, useCallback, useRef } from 'react'
import { Search as SearchIcon, X, User, Building2, Bell, Loader2 } from 'lucide-react'
import { useDebounce } from '../hooks/useDebounce'
import { api } from '../lib/api'

const EntityIcon = ({ type }) => {
  switch (type) {
    case 'user':
      return <User className="w-4 h-4" />
    case 'organization':
      return <Building2 className="w-4 h-4" />
    case 'notification':
      return <Bell className="w-4 h-4" />
    default:
      return <SearchIcon className="w-4 h-4" />
  }
}

const EntityBadge = ({ type }) => {
  const colors = {
    user: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
    organization: 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300',
    notification: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  }

  return (
    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${colors[type] || 'bg-gray-100 text-gray-700'}`}>
      <EntityIcon type={type} />
      {type}
    </span>
  )
}

const SearchResult = ({ result, onClick, isHighlighted }) => {
  return (
    <button
      onClick={() => onClick(result)}
      className={`w-full px-4 py-3 text-left hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors ${
        isHighlighted ? 'bg-gray-50 dark:bg-gray-700' : ''
      }`}
    >
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 mt-1">
          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white">
            <EntityIcon type={result.entity_type} />
          </div>
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <h4 className="text-sm font-medium text-gray-900 dark:text-white truncate">
              {result.title}
            </h4>
            <EntityBadge type={result.entity_type} />
          </div>

          {result.description && (
            <p className="text-xs text-gray-600 dark:text-gray-400 line-clamp-2">
              {result.highlight || result.description}
            </p>
          )}

          <div className="flex items-center gap-3 mt-2">
            <span className="text-xs text-gray-500 dark:text-gray-500">
              Relevance: {(result.rank * 100).toFixed(0)}%
            </span>
            <span className="text-xs text-gray-500 dark:text-gray-500">
              {new Date(result.created_at).toLocaleDateString()}
            </span>
          </div>
        </div>
      </div>
    </button>
  )
}

export default function Search({ onResultClick, placeholder = 'Search...', autoFocus = false }) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState(null)
  const [highlightedIndex, setHighlightedIndex] = useState(-1)

  const searchRef = useRef(null)
  const inputRef = useRef(null)

  const debouncedQuery = useDebounce(query, 300)

  // Handle click outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (searchRef.current && !searchRef.current.contains(event.target)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // Perform search
  const performSearch = useCallback(async (searchQuery) => {
    if (!searchQuery.trim()) {
      setResults([])
      setIsOpen(false)
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      const response = await api.get('/search', {
        params: {
          q: searchQuery,
          page: 1,
          page_size: 10,
        },
      })

      setResults(response.data.results || [])
      setIsOpen(true)
      setHighlightedIndex(-1)
    } catch (err) {
      console.error('Search error:', err)
      setError('Failed to perform search')
      setResults([])
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    performSearch(debouncedQuery)
  }, [debouncedQuery, performSearch])

  // Handle keyboard navigation
  const handleKeyDown = (e) => {
    if (!isOpen) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setHighlightedIndex((prev) => (prev < results.length - 1 ? prev + 1 : prev))
        break
      case 'ArrowUp':
        e.preventDefault()
        setHighlightedIndex((prev) => (prev > 0 ? prev - 1 : -1))
        break
      case 'Enter':
        e.preventDefault()
        if (highlightedIndex >= 0 && highlightedIndex < results.length) {
          handleResultClick(results[highlightedIndex])
        }
        break
      case 'Escape':
        e.preventDefault()
        setIsOpen(false)
        inputRef.current?.blur()
        break
    }
  }

  const handleResultClick = (result) => {
    setIsOpen(false)
    setQuery('')
    if (onResultClick) {
      onResultClick(result)
    }
  }

  const handleClear = () => {
    setQuery('')
    setResults([])
    setIsOpen(false)
    inputRef.current?.focus()
  }

  return (
    <div ref={searchRef} className="relative w-full max-w-2xl">
      {/* Search Input */}
      <div className="relative">
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          {isLoading ? (
            <Loader2 className="w-5 h-5 text-gray-400 animate-spin" />
          ) : (
            <SearchIcon className="w-5 h-5 text-gray-400" />
          )}
        </div>

        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          onFocus={() => {
            if (results.length > 0) {
              setIsOpen(true)
            }
          }}
          placeholder={placeholder}
          autoFocus={autoFocus}
          className="w-full pl-10 pr-10 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />

        {query && (
          <button
            onClick={handleClear}
            className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          >
            <X className="w-5 h-5" />
          </button>
        )}
      </div>

      {/* Search Results Dropdown */}
      {isOpen && (
        <div className="absolute z-50 w-full mt-2 bg-white dark:bg-gray-800 rounded-lg shadow-2xl border border-gray-200 dark:border-gray-700 max-h-96 overflow-y-auto">
          {error && (
            <div className="px-4 py-3 text-sm text-red-600 dark:text-red-400">
              {error}
            </div>
          )}

          {!error && results.length === 0 && query && !isLoading && (
            <div className="px-4 py-8 text-center">
              <SearchIcon className="w-12 h-12 mx-auto text-gray-400 mb-3" />
              <p className="text-sm text-gray-600 dark:text-gray-400">
                No results found for "{query}"
              </p>
            </div>
          )}

          {!error && results.length > 0 && (
            <>
              <div className="px-4 py-2 border-b border-gray-200 dark:border-gray-700">
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  {results.length} result{results.length !== 1 ? 's' : ''}
                </p>
              </div>

              <div className="divide-y divide-gray-200 dark:divide-gray-700">
                {results.map((result, index) => (
                  <SearchResult
                    key={`${result.entity_type}-${result.entity_id}`}
                    result={result}
                    onClick={handleResultClick}
                    isHighlighted={index === highlightedIndex}
                  />
                ))}
              </div>
            </>
          )}
        </div>
      )}

      {/* Keyboard Shortcuts Hint */}
      {isOpen && results.length > 0 && (
        <div className="absolute z-50 w-full mt-1 px-3 py-2 bg-gray-50 dark:bg-gray-900 rounded-lg text-xs text-gray-500 dark:text-gray-400 flex items-center justify-between">
          <span>↑↓ to navigate</span>
          <span>Enter to select</span>
          <span>Esc to close</span>
        </div>
      )}
    </div>
  )
}
