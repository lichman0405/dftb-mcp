import { createClient } from 'redis';
import { createHash } from 'crypto';
import { DFTBOptimizationRequest, DFTBOptimizationResponse, CacheConfig, CacheStats, CacheEntry } from '../types';

export class CacheService {
  private client: any;
  private config: CacheConfig;
  private stats: CacheStats;

  constructor(config: CacheConfig) {
    this.config = config;
    this.stats = {
      hits: 0,
      misses: 0,
      size: 0
    };
    
    if (config.enabled) {
      this.initializeRedis();
    }
  }

  /**
   * Initialize Redis client
   */
  private async initializeRedis(): Promise<void> {
    try {
      this.client = createClient({
        url: this.config.url,
        socket: {
          reconnectStrategy: (retries: number) => {
            const delay = Math.min(retries * 50, 1000);
            return delay;
          }
        }
      });

      this.client.on('error', (err: Error) => {
        console.error('Redis Client Error:', err);
      });

      this.client.on('connect', () => {
        console.log('Connected to Redis cache');
      });

      await this.client.connect();
    } catch (error) {
      console.error('Failed to initialize Redis:', error);
      this.config.enabled = false;
    }
  }

  /**
   * Generate cache key from request parameters
   */
  private generateCacheKey(request: DFTBOptimizationRequest): string {
    const keyData = {
      structure_file: request.structure_file,
      method: request.method || 'GFN1-xTB',
      fmax: request.fmax || 0.1
    };
    
    const keyString = JSON.stringify(keyData);
    const hash = createHash('md5').update(keyString).digest('hex');
    
    return `${this.config.prefix}:${hash}`;
  }

  /**
   * Get cached response
   */
  async get(request: DFTBOptimizationRequest): Promise<DFTBOptimizationResponse | null> {
    if (!this.config.enabled || !this.client) {
      return null;
    }

    try {
      const key = this.generateCacheKey(request);
      const cached = await this.client.get(key);
      
      if (cached) {
        const cacheEntry: CacheEntry = JSON.parse(cached);
        
        // Check if cache entry is expired
        if (Date.now() - cacheEntry.timestamp > cacheEntry.ttl * 1000) {
          await this.client.del(key);
          this.stats.misses++;
          return null;
        }
        
        this.stats.hits++;
        console.log(`Cache hit for key: ${key}`);
        return cacheEntry.value;
      }
      
      this.stats.misses++;
      return null;
    } catch (error) {
      console.error('Cache get error:', error);
      this.stats.misses++;
      return null;
    }
  }

  /**
   * Set cached response
   */
  async set(request: DFTBOptimizationRequest, response: DFTBOptimizationResponse): Promise<void> {
    if (!this.config.enabled || !this.client) {
      return;
    }

    try {
      const key = this.generateCacheKey(request);
      const cacheEntry: CacheEntry = {
        key,
        value: response,
        timestamp: Date.now(),
        ttl: this.config.ttl
      };
      
      await this.client.setEx(key, this.config.ttl, JSON.stringify(cacheEntry));
      console.log(`Cached response for key: ${key}`);
      
      // Update cache size
      this.stats.size = await this.getCacheSize();
    } catch (error) {
      console.error('Cache set error:', error);
    }
  }

  /**
   * Delete cached response
   */
  async delete(request: DFTBOptimizationRequest): Promise<boolean> {
    if (!this.config.enabled || !this.client) {
      return false;
    }

    try {
      const key = this.generateCacheKey(request);
      const result = await this.client.del(key);
      this.stats.size = await this.getCacheSize();
      return result > 0;
    } catch (error) {
      console.error('Cache delete error:', error);
      return false;
    }
  }

  /**
   * Clear all cache entries with the current prefix
   */
  async clear(): Promise<boolean> {
    if (!this.config.enabled || !this.client) {
      return false;
    }

    try {
      const keys = await this.client.keys(`${this.config.prefix}:*`);
      if (keys.length > 0) {
        await this.client.del(keys);
      }
      
      this.stats.size = 0;
      console.log(`Cleared ${keys.length} cache entries`);
      return true;
    } catch (error) {
      console.error('Cache clear error:', error);
      return false;
    }
  }

  /**
   * Get cache statistics
   */
  getStats(): CacheStats {
    return { ...this.stats };
  }

  /**
   * Get cache size
   */
  private async getCacheSize(): Promise<number> {
    if (!this.config.enabled || !this.client) {
      return 0;
    }

    try {
      const keys = await this.client.keys(`${this.config.prefix}:*`);
      return keys.length;
    } catch (error) {
      console.error('Cache size error:', error);
      return 0;
    }
  }

  /**
   * Check if cache is enabled
   */
  isEnabled(): boolean {
    return this.config.enabled && this.client !== null;
  }

  /**
   * Close Redis connection
   */
  async close(): Promise<void> {
    if (this.client) {
      await this.client.quit();
      console.log('Redis connection closed');
    }
  }
}
