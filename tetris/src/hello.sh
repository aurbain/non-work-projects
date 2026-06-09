#!/bin/bash
LIMIT=1000000

# Initialize an array to track primality. 1 means potentially prime, 0 means composite.
# We use an associative array or a simple map structure for better handling,
# but for simplicity with basic bash arrays, we'll use an indexed array.
# Since bash arrays are tricky for dynamic lookups, we'll use an auxiliary file or
# rely on direct checks if the structure is too complex for basic shell arrays.

# Let's simplify the structure to avoid complex array logic that clashes with shell syntax.
# We will use a standard Sieve implementation structure.

# Initialize an array (pseudo-Boolean: 1 for prime, 0 for composite)
declare -a isPrime
for ((i=0; i<=LIMIT; i++)); do
    isPrime[$i]=1
done

# 0 and 1 are not prime
isPrime[0]=0
isPrime[1]=0

p=2
while [ $p -le $LIMIT ]; do
    # Check if p is prime (isPrime[p] == 1)
    if [ "${isPrime[$p]}" -eq 1 ]; then
        echo "$p is prime" # Output the prime number
        # Mark all multiples of p starting from p*p as composite (set to 0)
        for ((i=p*p; i<=LIMIT; i+=p )); do
            isPrime[$i]=0
        done
    fi
    ((p++))
done

