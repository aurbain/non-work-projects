#!/bin/bash

OUTPUT_COUNT=${1:-100}
BUFFER_SIZE=$(( (OUTPUT_COUNT * 100) + 50 ))
declare -a is_prime
is_prime=()

for (( i=0; i<(BUFFER_SIZE-3)/2; i++ )); do
    is_prime[$i]=1
done

for (( i=0; i<(BUFFER_SIZE-3)/2; i++ )); do
    num=$(( 2 * i + 3 ))
    if [ "${is_prime[$i]}" -eq 1 ]; then
        echo "$num"
        for (( j=0; ; j++ )); do
            multiple=$(( num + j * num ))
            if [ $multiple -gt $BUFFER_SIZE ]; then break; fi
            multiple_idx=$(( (multiple - 3) / 2 ))
            if [ $multiple_idx -lt $(( (BUFFER_SIZE - 3) / 2 )) ] && [ $multiple_idx -gt 0 ]; then
                is_prime[$multiple_idx]=0
            fi
        done
    fi
done

echo "Done. Total primes found: $(echo $OUTPUT_COUNT)"
