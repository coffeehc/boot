package commons


const (
	key_const  = 2862933555777941757
	jump_const = float64(1 << 31)
)

/** 一致性 hash */
func JumpConsistentHash(key uint64, partition int64) int64 {
	var b int64 = -1
	var j int64 = 0
	for j < partition {
		b = j
		key = key*key_const + 1
		j = int64(float64(b+1) * jump_const / float64((key>>33)+1))
	}
	return b

}
