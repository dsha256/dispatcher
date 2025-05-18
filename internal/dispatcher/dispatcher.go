package dispatcher

import (
	"context"
	"errors"
	"sort"
)

var (
	ErrMultipleSameDestination = errors.New("multiple same destination")
	ErrCycleInItinerary        = errors.New("cycle in itinerary")
	ErrDifferentStartingPoints = errors.New("different starting points")
)

type Dispatcher struct{}

func New() *Dispatcher {
	return &Dispatcher{}
}

func (d *Dispatcher) ReconstructItinerary(_ context.Context, tickets *[][]string) ([]string, error) {
	return ReconstructItinerary(*tickets)
}

// ReconstructItinerary reconstructs a valid flight itinerary from a list of airline tickets.
// It uses a modified version of Hierholzer's algorithm to find a valid path that visits all destinations exactly once.
//
// Parameters:
//   - tickets: A slice of string pairs where each pair represents a flight ticket [from, to] A.K.A ["Source","Destination"].
//
// Returns:
//   - []string: The reconstructed itinerary as a sequence of airports
//   - error: Error if the itinerary is invalid
//
// Possible errors:
//   - ErrMultipleSameDestination: When there are multiple tickets with the same source and destination
//   - ErrCycleInItinerary: When the itinerary forms a cycle
//   - ErrDifferentStartingPoints: When there are multiple valid starting points or invalid graph structure
//
// Algorithm modifications from classical Hierholzer's:
// 1. Ensures no duplicate edges (tickets) are allowed
// 2. Prevents cycles in the final path
// 3. Validates proper start/end points before path finding
// 4. Uses lexicographically larger destinations first (reversed sort).
func ReconstructItinerary(tickets [][]string) ([]string, error) {
	if len(tickets) == 0 {
		return []string{}, nil
	}

	if _, err := validateTickets(tickets); err != nil {
		return nil, err
	}

	graph, outDegree, inDegree := buildGraph(tickets)

	start, err := findStartingPoint(outDegree, inDegree)
	if err != nil {
		return nil, err
	}

	startCandidates := []string{start}
	if err := validateEndPoints(startCandidates, outDegree, inDegree); err != nil {
		return nil, err
	}

	result := findPath(start, graph)

	if len(result) >= 2 && result[0] == result[len(result)-1] {
		return nil, ErrCycleInItinerary
	}

	return result, nil
}

// validateTickets checks for duplicate tickets and returns a map of ticket counts.
func validateTickets(tickets [][]string) (map[[2]string]int, error) {
	ticketCount := make(map[[2]string]int)
	for _, ticket := range tickets {
		key := [2]string{ticket[0], ticket[1]}
		ticketCount[key]++
		if ticketCount[key] > 1 {
			return nil, ErrMultipleSameDestination
		}
	}

	return ticketCount, nil
}

// buildGraph creates adjacency list and degree maps from tickets.
func buildGraph(tickets [][]string) (map[string][]string, map[string]int, map[string]int) {
	graph := make(map[string][]string)
	outDegree := make(map[string]int)
	inDegree := make(map[string]int)

	for _, ticket := range tickets {
		src, dst := ticket[0], ticket[1]
		graph[src] = append(graph[src], dst)
		outDegree[src]++
		inDegree[dst]++
	}

	for src := range graph {
		sort.Slice(graph[src], func(i, j int) bool {
			return graph[src][i] > graph[src][j]
		})
	}

	return graph, outDegree, inDegree
}

// findStartingPoint determines the valid starting airport.
func findStartingPoint(outDegree, inDegree map[string]int) (string, error) {
	startCandidates := []string{}
	validStart := true

	for node := range outDegree {
		diff := outDegree[node] - inDegree[node]
		if diff == 1 {
			startCandidates = append(startCandidates, node)
		} else if diff < -1 || diff > 1 {
			validStart = false

			break
		}
	}

	if !validStart || len(startCandidates) > 1 {
		return "", ErrDifferentStartingPoints
	}

	if len(startCandidates) == 1 {
		return startCandidates[0], nil
	}

	// Check if all nodes are balanced
	for node := range outDegree {
		if outDegree[node] != inDegree[node] {
			return "", ErrDifferentStartingPoints
		}
	}

	return "", ErrDifferentStartingPoints
}

// validateEndPoints ensures the graph has valid end points.
func validateEndPoints(startCandidates []string, outDegree, inDegree map[string]int) error {
	endCandidates := 0
	for node := range inDegree {
		diff := inDegree[node] - outDegree[node]
		if diff == 1 {
			endCandidates++
		} else if diff < -1 || diff > 1 {
			return ErrDifferentStartingPoints
		}
	}
	if len(startCandidates) == 1 && endCandidates != 1 {
		return ErrDifferentStartingPoints
	}

	return nil
}

// findPath uses modified Hierholzer's algorithm to find the path.
func findPath(start string, graph map[string][]string) []string {
	var result []string
	stack := []string{start}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]

		if dests, exists := graph[curr]; exists && len(dests) > 0 {
			nextDest := dests[len(dests)-1]
			graph[curr] = dests[:len(dests)-1]
			stack = append(stack, nextDest)
		} else {
			result = append(result, curr)
			stack = stack[:len(stack)-1]
		}
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}
