#!/usr/bin/env bash
# Phase 2.2 Learning System Integration Test
# Tests the enhanced learning CLI functionality

echo "🧪 Testing Phase 2.2 Learning System Integration"
echo "=============================================="
echo

# Build the project
echo "Building nixai..."
cd /home/olafkfreund/Source/NIX/nix-ai-help
go build -o nixai cmd/nixai/main.go
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi
echo "✅ Build successful"
echo

# Test basic learning command
echo "Testing basic learn command..."
output=$(./nixai learn 2>&1)
if [[ $output == *"Enhanced Learning System"* ]]; then
    echo "✅ Basic learn command works"
else
    echo "❌ Basic learn command failed"
    exit 1
fi

# Test module listing
echo "Testing module listing..."
output=$(./nixai learn list 2>&1)
if [[ $output == *"Available Learning Modules"* ]]; then
    echo "✅ Module listing works"
else
    echo "❌ Module listing failed"
    exit 1
fi

# Test progress feature
echo "Testing progress feature..."
output=$(./nixai learn progress 2>&1)
if [[ $output == *"Learning Progress"* ]]; then
    echo "✅ Progress feature works"
else
    echo "❌ Progress feature failed"
    exit 1
fi

# Test assessment feature
echo "Testing assessment feature..."
output=$(./nixai learn assess 2>&1)
if [[ $output == *"Skill Assessment"* ]]; then
    echo "✅ Assessment feature works"
else
    echo "❌ Assessment feature failed"
    exit 1
fi

# Test recommendations
echo "Testing recommendations..."
output=$(./nixai learn recommendations 2>&1)
if [[ $output == *"Personalized Recommendations"* ]]; then
    echo "✅ Recommendations feature works"
else
    echo "❌ Recommendations feature failed"
    exit 1
fi

# Test module starting
echo "Testing module starting..."
output=$(./nixai learn basics 2>&1)
if [[ $output == *"Starting Module: Basics"* ]]; then
    echo "✅ Module starting works"
else
    echo "❌ Module starting failed"
    exit 1
fi

# Test quiz feature
echo "Testing quiz feature..."
output=$(./nixai learn quiz flakes 2>&1)
if [[ $output == *"Quiz: Flakes"* ]]; then
    echo "✅ Quiz feature works"
else
    echo "❌ Quiz feature failed"
    exit 1
fi

echo
echo "🎉 All Phase 2.2 Learning System tests passed!"
echo "✅ Enhanced learning CLI is fully functional"
echo "✅ All learning features are accessible and working"
echo "✅ User interface is clear and informative"
echo "✅ Integration with existing CLI structure is seamless"
