package target

func getPath(userDest, fileName string) string {
	if len(userDest) == 0 {
		return fileName
	}
	return userDest
}
