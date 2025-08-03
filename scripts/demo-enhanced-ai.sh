#!/bin/bash

# Demo script for nixai's advanced AI capabilities

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

echo -e "${PURPLE}🚀 NixAI Advanced AI Demo${NC}"
echo -e "${CYAN}=========================${NC}"
echo ""

echo -e "${YELLOW}This demo showcases nixai's enhanced AI capabilities including:${NC}"
echo -e "${YELLOW}- Chain-of-thought reasoning${NC}"
echo -e "${YELLOW}- Self-correction mechanisms${NC}"
echo -e "${YELLOW}- Multi-step task planning${NC}"
echo -e "${YELLOW}- Confidence scoring${NC}"
echo ""

# Check if nixai is available
if ! command -v nixai &> /dev/null; then
    echo -e "${RED}nixai command not found!${NC}"
    echo -e "${YELLOW}Please install nixai first:${NC}"
    echo -e "${CYAN}  nix profile install github:olafkfreund/nix-ai-help${NC}"
    echo ""
    exit 1
fi

echo -e "${GREEN}nixai is available, proceeding with demo...${NC}"
echo ""

# Demo 1: Basic question with enhanced AI
echo -e "${BLUE}📝 Demo 1: Basic question with enhanced AI${NC}"
echo -e "${CYAN}-------------------------------------------${NC}"
echo -e "${YELLOW}Question: How to configure nginx in NixOS?${NC}"
echo ""
nixai -a "How to configure nginx in NixOS?"
echo ""
echo -e "${GREEN}✅ Done${NC}"
echo ""

# Demo 2: Complex task with planning
echo -e "${BLUE}📝 Demo 2: Complex task with planning${NC}"
echo -e "${CYAN}-------------------------------------${NC}"
echo -e "${YELLOW}Question: How to set up a development environment for Python and Django?${NC}"
echo ""
nixai -a "How to set up a development environment for Python and Django?"
echo ""
echo -e "${GREEN}✅ Done${NC}"
echo ""

# Demo 3: Self-correction in action
echo -e "${BLUE}📝 Demo 3: Self-correction in action${NC}"
echo -e "${CYAN}------------------------------------${NC}"
echo -e "${YELLOW}Question: What is the difference between flakes and channels?${NC}"
echo ""
nixai -a "What is the difference between flakes and channels?"
echo ""
echo -e "${GREEN}✅ Done${NC}"
echo ""

# Demo 4: Confidence scoring
echo -e "${BLUE}📝 Demo 4: Confidence scoring${NC}"
echo -e "${CYAN}-----------------------------${NC}"
echo -e "${YELLOW}Question: How to enable SSH?${NC}"
echo ""
nixai -a "How to enable SSH?"
echo ""
echo -e "${GREEN}✅ Done${NC}"
echo ""

# Demo 5: Edge case with self-correction
echo -e "${BLUE}📝 Demo 5: Edge case with self-correction${NC}"
echo -e "${CYAN}-----------------------------------------${NC}"
echo -e "${YELLOW}Question: (empty)${NC}"
echo ""
nixai -a ""
echo ""
echo -e "${GREEN}✅ Done${NC}"
echo ""

echo -e "${PURPLE}🎉 Demo completed!${NC}"
echo -e "${CYAN}==================${NC}"
echo ""
echo -e "${YELLOW}All demonstrations show nixai's advanced AI capabilities in action.${NC}"
echo -e "${YELLOW}Notice the transparent reasoning process, self-correction,${NC}"
echo -e "${YELLOW}task planning, and confidence scoring in each response.${NC}"
echo ""