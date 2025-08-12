# LLM Integration Examples for DFTB+ Optimization MCP Server

This document provides examples and guidance for integrating the DFTB+ Optimization MCP Server with various LLM platforms including Gemini CLI, Claude Code, and Cline.

## Table of Contents
- [Deployment Configuration](#deployment-configuration)
- [Gemini CLI Integration](#gemini-cli-integration)
- [Claude Code Integration](#claude-code-integration)
- [Cline Integration](#cline-integration)
- [LLM Tool Understanding](#llm-tool-understanding)
- [Prompt Engineering](#prompt-engineering)
- [Troubleshooting](#troubleshooting)

## Deployment Configuration

### Environment Variables

The MCP server supports flexible configuration through environment variables:

```bash
# Server Configuration
NODE_ENV=production
PORT=3000
GO_SERVICE_URL=http://localhost:8080

# Cache Configuration (Optional)
ENABLE_CACHE=true
REDIS_URL=redis://localhost:6379
CACHE_TTL=86400

# WebSocket Configuration
ENABLE_WEBSOCKET=true
MAX_REQUEST_SIZE=10485760
```

### Docker Deployment

```bash
# Start with Redis cache
docker-compose --profile redis up -d

# Start with monitoring
docker-compose --profile monitoring,redis up -d

# Full stack with nginx
docker-compose --profile nginx,redis,monitoring up -d
```

## Gemini CLI Integration

### Configuration File

Create or update your Gemini CLI configuration file (`~/.config/gemini/config.json`):

```json
{
  "mcp_servers": {
    "dftbopt": {
      "command": "node",
      "args": ["dist/server.js"],
      "env": {
        "NODE_ENV": "production",
        "PORT": "3000",
        "GO_SERVICE_URL": "http://localhost:8080",
        "REDIS_URL": "redis://localhost:6379",
        "ENABLE_CACHE": "true"
      }
    }
  }
}
```

### Usage Examples

#### Basic Optimization
```bash
gemini "Optimize this benzene molecule structure using GFN2-xTB with fmax=0.01"
```

#### Advanced Configuration
```bash
gemini "I need to optimize a crystal structure. Use the dftb_optimize tool with GFN1-xTB method and fmax=0.1. The structure is in CIF format."
```

### System Prompt Integration

Add this to your Gemini CLI system prompt:

```
You have access to a DFTB+ optimization tool that can perform quantum chemistry calculations on crystal structures.

Tool: dftb_optimize
Purpose: Perform geometry optimization on crystal structures using DFTB+ quantum chemistry package

When to use:
- User asks for molecular structure optimization
- User mentions geometry optimization
- User wants to minimize energy of a structure
- User needs to optimize crystal structures

Parameters:
- structure_file: Base64 encoded CIF content (required)
- method: "GFN1-xTB" (faster) or "GFN2-xTB" (more accurate)
- fmax: Force convergence threshold (0.001-0.1 eV/Angstrom)
- original_filename: Original filename for logging

Best practices:
- GFN1-xTB is suitable for most systems and faster
- GFN2-xTB provides better accuracy for certain elements
- Lower fmax values give more accurate results but take longer
- Always validate CIF format before submission
```

## Claude Code Integration

### MCP Configuration

Claude Code has native MCP support. Add this to your Claude Code MCP configuration (`~/.config/claude/mcp.json`):

```json
{
  "mcpServers": {
    "dftbopt": {
      "command": "node",
      "args": ["dist/server.js"],
      "env": {
        "NODE_ENV": "production",
        "PORT": "3000",
        "GO_SERVICE_URL": "http://localhost:8080",
        "REDIS_URL": "redis://localhost:6379",
        "ENABLE_CACHE": "true"
      }
    }
  }
}
```

### Tool Registration

Claude Code will automatically discover and register tools from the MCP server. The tool will be available as `dftb_optimize`.

### Usage Examples

#### Direct Tool Call
```javascript
// In Claude Code environment
const result = await tools.dftb_optimize({
  structure_file: "base64-encoded-cif-content",
  method: "GFN2-xTB",
  fmax: 0.01,
  original_filename: "benzene.cif"
});
```

#### Natural Language
```
Please optimize this crystal structure using the DFTB+ tool with GFN2-xTB method.
```

### System Prompt Enhancement

Claude Code benefits from detailed tool descriptions. The MCP server automatically provides rich metadata that Claude uses to understand the tool.

## Cline Integration

### Plugin Configuration

Cline supports MCP through its plugin system. Create a plugin configuration:

```yaml
# ~/.config/cline/plugins/dftbopt.yaml
name: dftbopt
version: "1.0.0"
type: mcp
config:
  command: node
  args: 
    - dist/server.js
  env:
    NODE_ENV: production
    PORT: "3000"
    GO_SERVICE_URL: "http://localhost:8080"
    REDIS_URL: "redis://localhost:6379"
    ENABLE_CACHE: "true"
```

### Usage Examples

#### Configuration File
```json
{
  "plugins": ["dftbopt"],
  "tools": {
    "dftb_optimize": {
      "description": "Perform DFTB+ geometry optimization",
      "endpoint": "http://localhost:3000/api/v1/optimize"
    }
  }
}
```

#### API Call
```bash
curl -X POST http://localhost:3000/api/v1/optimize \
  -H "Content-Type: application/json" \
  -d '{
    "structure_file": "base64-encoded-cif-content",
    "method": "GFN2-xTB",
    "fmax": 0.01,
    "original_filename": "benzene.cif"
  }'
```

## LLM Tool Understanding

### Automatic Tool Discovery

LLMs automatically discover MCP tools through the `tools/list` method. The server provides:

```json
{
  "tools": [
    {
      "name": "dftb_optimize",
      "description": "Perform DFTB+ geometry optimization on crystal structures. This tool takes a crystal structure file in CIF format and performs geometry optimization using the DFTB+ quantum chemistry package.",
      "inputSchema": {
        "type": "object",
        "properties": {
          "structure_file": {
            "type": "string",
            "description": "Base64 encoded content of the CIF structure file to optimize"
          },
          "method": {
            "type": "string",
            "enum": ["GFN1-xTB", "GFN2-xTB"],
            "default": "GFN1-xTB",
            "description": "The semi-empirical method to use for the calculation"
          },
          "fmax": {
            "type": "number",
            "default": 0.1,
            "description": "Force convergence threshold in eV/Angstrom"
          },
          "original_filename": {
            "type": "string",
            "description": "Optional original filename for the structure"
          }
        },
        "required": ["structure_file"]
      }
    }
  ]
}
```

### Tool Usage Patterns

LLMs learn to use the tool through:

1. **Description**: Clear, detailed tool description
2. **Schema**: Well-defined input parameters with types and constraints
3. **Examples**: Usage patterns embedded in descriptions
4. **Context**: When the tool should be used

### Prompt Strategy

#### Hybrid Approach (Recommended)

**System Level Prompt:**
```
You have access to quantum chemistry calculation tools for molecular structure optimization.

Available tool: dftb_optimize
Use case: When users need to optimize molecular geometries, minimize energy, or perform quantum chemistry calculations on crystal structures.

Key concepts:
- Geometry optimization finds minimum energy configurations
- DFTB+ is a semi-empirical quantum chemistry method
- GFN1-xTB is fast, GFN2-xTB is more accurate
- Lower fmax values give more precise results
```

**Tool Level Definition:**
The tool definition provides detailed technical parameters, validation rules, and format requirements.

**Dynamic Learning:**
LLMs learn from actual usage patterns and results, improving their understanding over time.

## Prompt Engineering

### Effective Prompt Components

1. **Context Setting**
   ```
   I need to optimize a crystal structure for my research.
   ```

2. **Tool Specification**
   ```
   Please use the DFTB+ optimization tool with GFN2-xTB method.
   ```

3. **Parameter Guidance**
   ```
   Use a tight convergence criterion (fmax=0.01) for high accuracy.
   ```

4. **Format Instructions**
   ```
   The structure is provided in CIF format below:
   ```

### Example Prompts

#### Basic Optimization
```
Optimize this benzene molecule using DFTB+ with GFN2-xTB method and fmax=0.01.

[CIF content here]
```

#### Advanced Configuration
```
I need to perform a geometry optimization on a protein-ligand complex. 
Please use the dftb_optimize tool with the following parameters:
- Method: GFN1-xTB (for faster computation)
- fmax: 0.05 (moderate accuracy)
- The structure is in CIF format

[CIF content here]
```

#### Comparative Study
```
Please optimize this structure using both GFN1-xTB and GFN2-xTB methods. 
Use fmax=0.01 for both calculations and compare the results.

[CIF content here]
```

### Error Handling

LLMs should be guided to handle common errors:

```
If the optimization fails, please:
1. Check if the CIF format is valid
2. Verify the structure contains atomic coordinates
3. Try with a larger fmax value (e.g., 0.1)
4. Report the specific error message
```

## Troubleshooting

### Common Issues

#### 1. Connection Problems
**Symptoms**: LLM cannot connect to MCP server
**Solutions**:
- Check if server is running: `curl http://localhost:3000/health`
- Verify network connectivity
- Check firewall settings
- Ensure correct port configuration

#### 2. Tool Not Found
**Symptoms**: LLM reports "tool not found"
**Solutions**:
- Verify MCP server is properly registered
- Check tool registration logs
- Ensure LLM platform supports MCP
- Restart LLM client after configuration changes

#### 3. Cache Issues
**Symptoms**: Inconsistent results or performance problems
**Solutions**:
- Check Redis connection: `redis-cli ping`
- Verify cache configuration
- Clear cache if needed
- Monitor cache hit/miss ratios

#### 4. Performance Issues
**Symptoms**: Slow response times
**Solutions**:
- Enable caching with Redis
- Use appropriate fmax values
- Monitor system resources
- Consider scaling horizontally

### Debug Commands

```bash
# Check server health
curl http://localhost:3000/health

# List available tools
curl http://localhost:3000/api/v1/tools

# Test optimization endpoint
curl -X POST http://localhost:3000/api/v1/optimize \
  -H "Content-Type: application/json" \
  -d '{"structure_file": "base64-content", "method": "GFN1-xTB"}'

# Check Redis status
redis-cli info

# View logs
docker-compose logs -f typescript-mcp
```

### Monitoring

Enable monitoring profile for comprehensive observability:

```bash
docker-compose --profile monitoring,redis up -d
```

Access:
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001 (admin/admin)

## Best Practices

### For LLM Users
1. **Provide Clear Context**: Explain what you want to optimize
2. **Specify Parameters**: Mention method and accuracy requirements
3. **Validate Input**: Ensure CIF format is correct
4. **Monitor Progress**: Check optimization status
5. **Handle Errors**: Understand common failure modes

### For System Administrators
1. **Use Environment Variables**: Configure all settings via environment
2. **Enable Caching**: Redis significantly improves performance
3. **Monitor Resources**: Track CPU, memory, and disk usage
4. **Scale Horizontally**: Use load balancers for multiple instances
5. **Implement Logging**: Centralized logging for debugging

### For Developers
1. **Test Thoroughly**: Validate with various CIF structures
2. **Handle Errors Gracefully**: Provide meaningful error messages
3. **Document APIs**: Keep OpenAPI specification updated
4. **Version Compatibility**: Ensure LLM platform compatibility
5. **Performance Optimize**: Profile and optimize critical paths

## Support

For issues and questions:
- GitHub Issues: [Project Repository]
- Documentation: [API Docs](http://localhost:3000/api-docs)
- Community: [Discussion Forum]

---

*This document will be updated as new LLM platforms and integration patterns emerge.*
