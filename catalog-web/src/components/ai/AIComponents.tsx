import React, { useState, useEffect, useCallback } from 'react';
import { Search, Brain, Lightbulb, Tag, Grid, Folder, Sparkles, TrendingUp } from 'lucide-react';

// Types for AI-powered suggestions
interface AISuggestion {
  id: string;
  title: string;
  description: string;
  type: 'collection' | 'smart-search' | 'content-categorization' | 'tag-suggestion';
  confidence: number;
  reasoning: string;
  metadata: {
    relatedItems: number;
    estimatedTimeSaved: string;
    category?: string;
    tags?: string[];
  };
}

interface AICategorizationResult {
  id: string;
  category: string;
  subcategory: string;
  confidence: number;
  reasoning: string;
  suggestedTags: string[];
  metadata: {
    contentType: string;
    quality: 'high' | 'medium' | 'low';
    completion: number;
  };
}

interface AISearchQuery {
  query: string;
  intent: 'browse' | 'search' | 'compare' | 'organize';
  entities: string[];
  filters: Record<string, any>;
  naturalLanguage: boolean;
}

// Mock AI service for demonstration
class AIService {
  // Simulate AI processing delay
  private static delay(ms = 800): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  // Generate collection suggestions based on user behavior and content patterns
  static async generateCollectionSuggestions(context: {
    recentSearches: string[];
    existingCollections: string[];
    contentAnalysis: any[];
    userBehavior: any;
  }): Promise<AISuggestion[]> {
    await this.delay(1000);

    const suggestions: AISuggestion[] = [
      {
        id: 'ai-collection-1',
        title: 'Movie Marathon Collection',
        description: 'Create a collection for movie marathon nights based on your recent viewing patterns',
        type: 'collection',
        confidence: 0.92,
        reasoning: 'You watched 5 action movies in the past 2 weeks. Marathon collections are 78% more engaging than individual items.',
        metadata: {
          relatedItems: 47,
          estimatedTimeSaved: '2.5 hours',
          category: 'Entertainment',
          tags: ['movies', 'marathon', 'action']
        }
      },
      {
        id: 'ai-collection-2',
        title: 'Work-from-Home Resources',
        description: 'Organize your work-related resources and productivity tools',
        type: 'collection',
        confidence: 0.88,
        reasoning: 'Recent searches show increased work-related content consumption. Productivity collections improve workflow by 45%.',
        metadata: {
          relatedItems: 23,
          estimatedTimeSaved: '1.8 hours',
          category: 'Work',
          tags: ['work', 'productivity', 'remote']
        }
      },
      {
        id: 'ai-collection-3',
        title: 'Learning Path: Web Development',
        description: 'Create a structured learning path from your development resources',
        type: 'collection',
        confidence: 0.85,
        reasoning: 'You have 12 development tutorials. Structured learning paths increase completion rates by 62%.',
        metadata: {
          relatedItems: 19,
          estimatedTimeSaved: '3.2 hours',
          category: 'Education',
          tags: ['development', 'learning', 'tutorial']
        }
      }
    ];

    return suggestions.filter(s => s.confidence > 0.8);
  }

  // Categorize content using AI analysis
  static async categorizeContent(item: {
    title: string;
    description?: string;
    metadata?: any;
  }): Promise<AICategorizationResult> {
    await this.delay(600);

    const categories = [
      'Entertainment', 'Work', 'Education', 'Personal', 'Health & Fitness', 
      'Finance', 'Technology', 'Travel', 'Food & Cooking', 'Shopping'
    ];

    const category = categories[Math.floor(Math.random() * categories.length)];
    const confidence = 0.75 + Math.random() * 0.2;

    return {
      id: `ai-cat-${Date.now()}`,
      category,
      subcategory: `${category}/${item.title.split(' ')[0]}`,
      confidence,
      reasoning: `AI analysis of title, content type, and metadata suggests ${category} category with ${Math.round(confidence * 100)}% confidence.`,
      suggestedTags: [category.toLowerCase(), item.title.toLowerCase().split(' ')[0], 'suggested'],
      metadata: {
        contentType: 'video',
        quality: confidence > 0.9 ? 'high' : confidence > 0.8 ? 'medium' : 'low',
        completion: Math.floor(Math.random() * 100)
      }
    };
  }

  // Process natural language search queries
  static async processNaturalLanguageQuery(query: string): Promise<AISearchQuery> {
    await this.delay(400);

    // Simple intent detection based on keywords
    const isBrowse = /show|browse|list|view|see/.test(query.toLowerCase());
    const isSearch = /find|search|look for|where is/.test(query.toLowerCase());
    const isCompare = /compare|difference|better|versus|vs/.test(query.toLowerCase());
    const isOrganize = /organize|group|sort|arrange/.test(query.toLowerCase());

    let intent: AISearchQuery['intent'] = 'search';
    if (isBrowse) intent = 'browse';
    else if (isCompare) intent = 'compare';
    else if (isOrganize) intent = 'organize';

    // Extract entities (simple keyword extraction)
    const entities = query
      .toLowerCase()
      .replace(/\b(the|a|an|and|or|but|in|on|at|to|for|of|with|by)\b/g, '')
      .split(' ')
      .filter(word => word.length > 2)
      .slice(0, 5);

    return {
      query,
      intent,
      entities,
      filters: {},
      naturalLanguage: true
    };
  }

  // Generate smart search suggestions
  static async generateSmartSearchSuggestions(partialQuery: string): Promise<string[]> {
    await this.delay(300);

    const suggestions = [
      `${partialQuery} collection`,
      `${partialQuery} high quality`,
      `recent ${partialQuery}`,
      `popular ${partialQuery}`,
      `${partialQuery} tutorial`,
      `${partialQuery} for beginners`,
      `complete ${partialQuery}`,
      `${partialQuery} review`
    ];

    return suggestions.slice(0, 5);
  }
}

// AI-powered collection suggestions component
interface AICollectionSuggestionsProps {
  onSuggestionAccept: (suggestion: AISuggestion) => void;
  maxSuggestions?: number;
}

export const AICollectionSuggestions: React.FC<AICollectionSuggestionsProps> = ({
  onSuggestionAccept,
  maxSuggestions = 3
}) => {
  const [suggestions, setSuggestions] = useState<AISuggestion[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadSuggestions = async () => {
      setLoading(true);
      setError(null);
      try {
        const result = await AIService.generateCollectionSuggestions({
          recentSearches: ['action movies', 'productivity tools', 'web development'],
          existingCollections: ['Favorites', 'Work Projects', 'Personal Growth'],
          contentAnalysis: [],
          userBehavior: {}
        });
        setSuggestions(result.slice(0, maxSuggestions));
      } catch (err) {
        setError('Failed to load AI suggestions');
      } finally {
        setLoading(false);
      }
    };

    loadSuggestions();
  }, [maxSuggestions]);

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        <div className="flex items-center space-x-3 mb-4">
          <Brain className="w-5 h-5 text-indigo-600" />
          <h3 className="font-semibold text-gray-900">AI Suggestions</h3>
          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-indigo-600 ml-auto"></div>
        </div>
        <div className="space-y-3">
          {[1, 2, 3].map(i => (
            <div key={i} className="animate-pulse">
              <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
              <div className="h-3 bg-gray-200 rounded w-full"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
        <div className="flex items-center space-x-2">
          <div className="w-5 h-5 text-red-600">⚠️</div>
          <span className="text-red-800">{error}</span>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
      <div className="flex items-center space-x-3 mb-4">
        <Brain className="w-5 h-5 text-indigo-600" />
        <h3 className="font-semibold text-gray-900">AI-Powered Suggestions</h3>
        <Sparkles className="w-4 h-4 text-yellow-500 ml-auto" />
      </div>

      {suggestions.length === 0 ? (
        <p className="text-gray-500 text-sm">No AI suggestions available at the moment.</p>
      ) : (
        <div className="space-y-3">
          {suggestions.map((suggestion) => (
            <div
              key={suggestion.id}
              className="border border-gray-100 rounded-lg p-3 hover:border-indigo-200 hover:bg-indigo-50 transition-colors cursor-pointer"
              onClick={() => onSuggestionAccept(suggestion)}
            >
              <div className="flex items-start justify-between mb-2">
                <div className="flex items-center space-x-2">
                  {suggestion.type === 'collection' && <Folder className="w-4 h-4 text-indigo-600" />}
                  <h4 className="font-medium text-gray-900">{suggestion.title}</h4>
                </div>
                <div className="flex items-center space-x-2">
                  <div className="flex items-center">
                    <div className="w-2 h-2 rounded-full bg-green-500 mr-1"></div>
                    <span className="text-xs text-gray-600">
                      {Math.round(suggestion.confidence * 100)}% match
                    </span>
                  </div>
                </div>
              </div>
              
              <p className="text-sm text-gray-600 mb-2">{suggestion.description}</p>
              
              <div className="flex items-center justify-between text-xs">
                <div className="flex items-center space-x-3 text-gray-500">
                  <span className="flex items-center space-x-1">
                    <Grid className="w-3 h-3" />
                    <span>{suggestion.metadata.relatedItems} items</span>
                  </span>
                  <span className="flex items-center space-x-1">
                    <TrendingUp className="w-3 h-3" />
                    <span>{suggestion.metadata.estimatedTimeSaved} saved</span>
                  </span>
                </div>
                {suggestion.metadata.tags && (
                  <div className="flex space-x-1">
                    {suggestion.metadata.tags.slice(0, 2).map(tag => (
                      <span key={tag} className="px-2 py-1 bg-gray-100 text-gray-600 rounded-full text-xs">
                        {tag}
                      </span>
                    ))}
                  </div>
                )}
              </div>
              
              <div className="mt-2 p-2 bg-blue-50 rounded text-xs text-blue-700">
                <strong>AI Reasoning:</strong> {suggestion.reasoning}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

// Natural language search component
interface AINaturalSearchProps {
  onSearch: (query: AISearchQuery) => void;
  placeholder?: string;
}

export const AINaturalSearch: React.FC<AINaturalSearchProps> = ({
  onSearch,
  placeholder = "Search naturally... try 'show me action movies' or 'find productivity tools'"
}) => {
  const [query, setQuery] = useState('');
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);

  const debouncedGetSuggestions = useCallback(
    debounce(async (input: string) => {
      if (input.length < 3) {
        setSuggestions([]);
        return;
      }

      setLoading(true);
      try {
        const result = await AIService.generateSmartSearchSuggestions(input);
        setSuggestions(result);
      } catch (err) {
        console.error('Failed to get suggestions:', err);
      } finally {
        setLoading(false);
      }
    }, 300),
    []
  );

  useEffect(() => {
    debouncedGetSuggestions(query);
  }, [query, debouncedGetSuggestions]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    setLoading(true);
    try {
      const processedQuery = await AIService.processNaturalLanguageQuery(query);
      onSearch(processedQuery);
      setShowSuggestions(false);
    } catch (err) {
      console.error('Failed to process query:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="relative">
      <form onSubmit={handleSubmit} className="relative">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onFocus={() => setShowSuggestions(true)}
            placeholder={placeholder}
            className="w-full pl-10 pr-12 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
          <div className="absolute right-3 top-1/2 transform -translate-y-1/2 flex items-center space-x-2">
            {loading && (
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-indigo-600"></div>
            )}
            <Brain className="w-5 h-5 text-indigo-600" />
          </div>
        </div>
      </form>

      {showSuggestions && suggestions.length > 0 && (
        <div className="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg">
          <div className="p-2 border-b border-gray-100">
            <div className="flex items-center space-x-2 text-xs text-gray-600">
              <Lightbulb className="w-3 h-3" />
              <span>AI-powered suggestions</span>
            </div>
          </div>
          {suggestions.map((suggestion, index) => (
            <button
              key={index}
              onClick={() => {
                setQuery(suggestion);
                setShowSuggestions(false);
              }}
              className="w-full text-left px-4 py-2 hover:bg-gray-50 text-sm text-gray-700 hover:text-indigo-600 transition-colors"
            >
              {suggestion}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

// Smart content categorization component
interface AIContentCategorizerProps {
  item: {
    title: string;
    description?: string;
    metadata?: any;
  };
  onCategorizationComplete: (result: AICategorizationResult) => void;
}

export const AIContentCategorizer: React.FC<AIContentCategorizerProps> = ({
  item,
  onCategorizationComplete
}) => {
  const [categorization, setCategorization] = useState<AICategorizationResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [categorizing, setCategorizing] = useState(false);

  const categorizeContent = async () => {
    setCategorizing(true);
    try {
      const result = await AIService.categorizeContent(item);
      setCategorization(result);
      onCategorizationComplete(result);
    } catch (err) {
      console.error('Failed to categorize content:', err);
    } finally {
      setCategorizing(false);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
      <div className="flex items-center space-x-3 mb-4">
        <Tag className="w-5 h-5 text-indigo-600" />
        <h3 className="font-semibold text-gray-900">Smart Content Categorization</h3>
      </div>

      {!categorization && !categorizing && (
        <div className="text-center py-4">
          <p className="text-gray-600 mb-4">AI can automatically categorize your content</p>
          <button
            onClick={categorizeContent}
            className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors flex items-center space-x-2 mx-auto"
          >
            <Brain className="w-4 h-4" />
            <span>Categorize with AI</span>
          </button>
        </div>
      )}

      {categorizing && (
        <div className="text-center py-4">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto mb-3"></div>
          <p className="text-gray-600">AI is analyzing your content...</p>
        </div>
      )}

      {categorization && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <h4 className="font-medium text-gray-900">{categorization.category}</h4>
              <p className="text-sm text-gray-600">{categorization.subcategory}</p>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 rounded-full bg-green-500"></div>
              <span className="text-sm text-gray-600">
                {Math.round(categorization.confidence * 100)}% confidence
              </span>
            </div>
          </div>

          <div className="p-3 bg-gray-50 rounded-lg">
            <p className="text-sm text-gray-700">{categorization.reasoning}</p>
          </div>

          <div>
            <h5 className="text-sm font-medium text-gray-700 mb-2">Suggested Tags:</h5>
            <div className="flex flex-wrap gap-2">
              {categorization.suggestedTags.map(tag => (
                <span key={tag} className="px-3 py-1 bg-indigo-100 text-indigo-700 rounded-full text-sm">
                  {tag}
                </span>
              ))}
            </div>
          </div>

          <button
            onClick={categorizeContent}
            className="text-sm text-indigo-600 hover:text-indigo-800 transition-colors"
          >
            Recategorize with AI
          </button>
        </div>
      )}
    </div>
  );
};

// Export AI service for external use
export { AIService };
export type { AISuggestion, AICategorizationResult, AISearchQuery };