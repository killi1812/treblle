#!/bin/bash

# --- Configuration ---
PROXY_URL="http://localhost:8090/proxy/api/json/v1/1" # Your proxy endpoint base
MIN_REQUESTS=50
MAX_REQUESTS=150
MIN_WAIT=0.1  # Minimum wait time in seconds (can be float)
MAX_WAIT=1.5  # Maximum wait time in seconds (can be float)

# --- Data for random requests ---
COCKTAIL_NAMES=("Margarita" "Mojito" "Martini" "Cosmopolitan" "Old Fashioned" "Bloody Mary" "Pina Colada" "Daiquiri" "Gin Tonic")
FIRST_LETTERS=({a..z})
INGREDIENTS=("Vodka" "Gin" "Rum" "Tequila" "Whiskey" "Orange Juice" "Lime" "Sugar" "Mint")
COCKTAIL_IDS=("11007" "17222" "11000" "17196" "11001" "11113" "17207" "11308" "11410") # Example IDs
INGREDIENT_IDS=("552" "2" "3" "17" "4" "279" "30" "41" "34") # Example IDs
ALCOHOLIC_FILTERS=("Alcoholic" "Non_Alcoholic")
CATEGORY_FILTERS=("Ordinary_Drink" "Cocktail" "Shot" "Punch / Party Drink")
GLASS_FILTERS=("Cocktail_glass" "Champagne_flute" "Highball glass" "Old-fashioned glass")
LIST_FILTERS=("c=list" "g=list" "i=list" "a=list")

# Function to generate a random float between min and max
random_float() {
  awk -v min="$1" -v max="$2" 'BEGIN{srand(); print min+rand()*(max-min)}'
}

# --- Determine total number of requests ---
TOTAL_REQUESTS=$(( RANDOM % (MAX_REQUESTS - MIN_REQUESTS + 1) + MIN_REQUESTS ))
echo "Generating $TOTAL_REQUESTS random requests..."
echo "Targeting proxy: $PROXY_URL"
echo "-------------------------------------"

# --- Main Loop ---
for (( i=1; i<=TOTAL_REQUESTS; i++ )); do
  echo -n "Request $i/$TOTAL_REQUESTS: "

  # Choose a random endpoint type
  ENDPOINT_TYPE=$(( RANDOM % 10 ))

  # Construct URL based on type
  case $ENDPOINT_TYPE in
    0) # Search by name
      RANDOM_NAME=$(printf "%s" "${COCKTAIL_NAMES[RANDOM % ${#COCKTAIL_NAMES[@]}]}" | sed 's/ /%20/g') # URL encode spaces
      API_PATH="/search.php?s=${RANDOM_NAME}"
      echo "Search Name: ${RANDOM_NAME}"
      ;;
    1) # List by first letter
      RANDOM_LETTER=${FIRST_LETTERS[RANDOM % ${#FIRST_LETTERS[@]}]}
      API_PATH="/search.php?f=${RANDOM_LETTER}"
      echo "List Letter: ${RANDOM_LETTER}"
      ;;
    2) # Search ingredient by name
      RANDOM_INGREDIENT=$(printf "%s" "${INGREDIENTS[RANDOM % ${#INGREDIENTS[@]}]}" | sed 's/ /%20/g')
      API_PATH="/search.php?i=${RANDOM_INGREDIENT}"
      echo "Search Ingredient: ${RANDOM_INGREDIENT}"
      ;;
    3) # Lookup cocktail by ID
      RANDOM_ID=${COCKTAIL_IDS[RANDOM % ${#COCKTAIL_IDS[@]}]}
      API_PATH="/lookup.php?i=${RANDOM_ID}"
      echo "Lookup Cocktail ID: ${RANDOM_ID}"
      ;;
    4) # Lookup ingredient by ID
      RANDOM_ID=${INGREDIENT_IDS[RANDOM % ${#INGREDIENT_IDS[@]}]}
      API_PATH="/lookup.php?iid=${RANDOM_ID}"
      echo "Lookup Ingredient ID: ${RANDOM_ID}"
      ;;
    5) # Lookup random cocktail
      API_PATH="/random.php"
      echo "Random Cocktail"
      ;;
    6) # Search by ingredient
      RANDOM_INGREDIENT=$(printf "%s" "${INGREDIENTS[RANDOM % ${#INGREDIENTS[@]}]}" | sed 's/ /%20/g')
      API_PATH="/filter.php?i=${RANDOM_INGREDIENT}"
      echo "Filter Ingredient: ${RANDOM_INGREDIENT}"
      ;;
    7) # Filter by alcoholic type
      RANDOM_FILTER=${ALCOHOLIC_FILTERS[RANDOM % ${#ALCOHOLIC_FILTERS[@]}]}
      API_PATH="/filter.php?a=${RANDOM_FILTER}"
      echo "Filter Alcoholic: ${RANDOM_FILTER}"
      ;;
    8) # Filter by Category or Glass
      if (( RANDOM % 2 == 0 )); then
        RANDOM_FILTER=$(printf "%s" "${CATEGORY_FILTERS[RANDOM % ${#CATEGORY_FILTERS[@]}]}" | sed 's/ /%20/g')
        API_PATH="/filter.php?c=${RANDOM_FILTER}"
        echo "Filter Category: ${RANDOM_FILTER}"
      else
        RANDOM_FILTER=$(printf "%s" "${GLASS_FILTERS[RANDOM % ${#GLASS_FILTERS[@]}]}" | sed 's/ /%20/g')
        API_PATH="/filter.php?g=${RANDOM_FILTER}"
        echo "Filter Glass: ${RANDOM_FILTER}"
      fi
      ;;
    9) # List filters
      RANDOM_FILTER=${LIST_FILTERS[RANDOM % ${#LIST_FILTERS[@]}]}
      API_PATH="/list.php?${RANDOM_FILTER}"
      echo "List Filter: ${RANDOM_FILTER}"
      ;;
  esac

  # Construct the full URL
  FULL_URL="${PROXY_URL}${API_PATH}"

  # Execute curl command (suppress output with -s -o /dev/null)
  # Remove '-s -o /dev/null' if you want to see the JSON output
  curl -s -o /dev/null -X GET "$FULL_URL" -H 'accept: application/json'

  # Wait a random amount of time
  WAIT_TIME=$(random_float $MIN_WAIT $MAX_WAIT)
  # Use sleep with floating point numbers
  sleep "$WAIT_TIME"

done

echo "-------------------------------------"
echo "Finished generating $TOTAL_REQUESTS requests."
