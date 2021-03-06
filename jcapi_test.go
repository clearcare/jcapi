package jcapi

import (
	"os"
	"testing"
)

const (
	testUrlBase string = "https://console.jumpcloud.com/api"
	authUrlBase string = "https://auth.jumpcloud.com"
)

var testAPIKey string = os.Getenv("JUMPCLOUD_APIKEY")

func MakeTestUser() (user JCUser) {
	user = JCUser{
		UserName:          "testuser",
		FirstName:         "Test",
		LastName:          "User",
		Email:             "testuser@jumpcloud.com",
		Password:          "test!@#$ADSF",
		Activated:         true,
		Sudo:              true,
		Uid:               "2244",
		Gid:               "2244",
		EnableManagedUid:  true,
		TagList:           make([]string, 0),
		ExternallyManaged: false,
	}

	return
}
func TestSystems(t *testing.T) {
	jcapi := NewJCAPI(testAPIKey, testUrlBase)
	systems, err := jcapi.GetSystems(true)
	if err != nil {
		t.Fatalf("couldn't get system")
	}
	if len(systems) == 0 {
		t.Fatalf("no systems found")
	}
	//fmt.Printf("'%d' Systems found\n", len(systems))
	testSystem := systems[0]
	sysByID, err := jcapi.GetSystemById(testSystem.Id, true)
	if testSystem.Id != sysByID.Id {
		t.Fatalf("Got ID='%s', expected '%s'", sysByID.Id, testSystem.Id)
	}
	t.Logf("TestSystem: '%s'", testSystem.ToString())
	var foundSystem int = -1
	sysByHostname, err := jcapi.GetSystemByHostName(testSystem.Hostname, true)
	for i, sys := range sysByHostname {
		if sys.Id == testSystem.Id {
			foundSystem = i
		}
	}
	if foundSystem == -1 {
		t.Fatalf("Didn't find test system '%s', foundSystem=%d", testSystem.Id, foundSystem)
	}
	t.Logf("TestSystem: '%s'", testSystem.ToString())
	tagsBefore := testSystem.Tags
	if len(tagsBefore) == 0 {
		t.Fatalf("no tags in test system :-(")
	}
	allTags, err := jcapi.GetAllTags()
	if err != nil {
		t.Fatalf("couldn't get the tags")
	}
	tagList := make([]string, len(allTags))
	for i, tag := range allTags {
		tagList[i] = tag.Name
	}
	testSystem.TagList = tagList
	updatedSystemId, err := jcapi.UpdateSystem(testSystem)
	if err != nil {
		t.Fatalf("Couldn't update system, err='%s'", err)
	}
	updatedSystem, err := jcapi.GetSystemById(updatedSystemId, true)
	if err != nil {
		t.Fatalf("error getting system")
	}
	tagsAfter := updatedSystem.Tags
	if len(tagsAfter) < len(allTags) {
		t.Fatalf("not enough tags!")
	}
	beforeTagList := make([]string, len(tagsBefore))
	for i, tag := range tagsBefore {
		beforeTagList[i] = tag.Name
	}

	updatedSystem.TagList = beforeTagList
	backToNormalId, err := jcapi.UpdateSystem(updatedSystem)
	if err != nil {
		t.Fatalf("Couldn't update system, err='%s'", err)
	}
	backToNormal, err := jcapi.GetSystemById(backToNormalId, true)
	if err != nil {
		t.Fatalf("error getting system")
	}
	//TODO complare Tags contents
	if len(backToNormal.Tags) != len(tagsBefore) {
		t.Fatalf("tags don't match")
	}

}

func TestSystemUsersByOne(t *testing.T) {
	jcapi := NewJCAPI(testAPIKey, testUrlBase)

	newUser := MakeTestUser()

	userId, err := jcapi.AddUpdateUser(Insert, newUser)
	if err != nil {
		t.Fatalf("Could not add new user ('%s'), err='%s'", newUser.ToString(), err)
	}

	t.Logf("Returned userId=%s", userId)

	retrievedUser, err := jcapi.GetSystemUserById(userId, true)
	if err != nil {
		t.Fatalf("Could not get the system user I just added, err='%s'", err)
	}

	if userId != retrievedUser.Id {
		t.Fatalf("Got back a different user ID than expected, this shouldn't happen! Initial userId='%s' - returned object: '%s'",
			userId, retrievedUser.ToString())
	}

	retrievedUser.Email = "newtestemail@jumpcloud.com"

	// We have to do the following because of bug: https://www.pivotaltracker.com/story/show/84876992
	retrievedUser.Uid = "2244"
	retrievedUser.Gid = "2244"

	newUserId, err := jcapi.AddUpdateUser(Update, retrievedUser)
	if err != nil {
		t.Fatalf("Could not modify email on the just-added user ('%s'), err='%s'", retrievedUser.ToString(), err)
	}

	if userId != newUserId {
		t.Fatalf("The user ID of the updated user changed across updates, this should never happen!")
	}

	err = jcapi.DeleteUser(retrievedUser)
	if err != nil {
		t.Fatalf("Could not delete user ('%s'), err='%s'", retrievedUser.ToString(), err)
	}

	return
}

func TestSystemUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping potentially long test in short mode")
	}

	jcapi := NewJCAPI(testAPIKey, testUrlBase)

	newUser := MakeTestUser()

	userId, err := jcapi.AddUpdateUser(Insert, newUser)
	if err != nil {
		t.Fatalf("Could not add new user ('%s'), err='%s'", newUser.ToString(), err)
	}

	t.Logf("Returned userId=%s", userId)

	allUsers, err := jcapi.GetSystemUsers(true)
	if err != nil {
		t.Fatalf("Could not get all system users, err='%s'", err)
	}

	t.Logf("GetSystemUsers() returned %d users", len(allUsers))

	var foundUser int = -1

	for i, user := range allUsers {
		if user.Id == userId {
			foundUser = i
			t.Logf("Matched user[%d]='%s'", i, user.ToString())
		}
	}

	if foundUser == -1 {
		t.Fatalf("Could not find the user ID just added '%s', foundUser=%d", userId, foundUser)
	}

	allUsers[foundUser].Email = "newtestemail@jumpcloud.com"

	// We have to do the following because of bug: https://www.pivotaltracker.com/story/show/84876992
	allUsers[foundUser].Uid = "2244"
	allUsers[foundUser].Gid = "2244"

	newUserId, err := jcapi.AddUpdateUser(Update, allUsers[foundUser])
	if err != nil {
		t.Fatalf("Could not modify email on the just-added user ('%s'), err='%s'", allUsers[foundUser].ToString(), err)
	}

	if userId != newUserId {
		t.Fatalf("The user ID of the updated user changed across updates, this should never happen!")
	}

	err = jcapi.DeleteUser(allUsers[foundUser])
	if err != nil {
		t.Fatalf("Could not delete user ('%s'), err='%s'", allUsers[foundUser].ToString(), err)
	}

	return
}

func MakeTestTag() (tag JCTag) {
	tag = JCTag{
		//Had to remove the '#' cause it was breaking
		//the unmarshal in the Do func and i'm
		//not sure how to fix it
		Name:              "Test tag 1",
		GroupName:         "testtag1",
		Systems:           make([]string, 0),
		SystemUsers:       make([]string, 0),
		Expired:           false,
		Selected:          false,
		ExternallyManaged: false,
	}

	return
}

func TestTags(t *testing.T) {
	jcapi := NewJCAPI(testAPIKey, testUrlBase)

	newTag := MakeTestTag()

	tagId, err := jcapi.AddUpdateTag(Insert, newTag)
	if err != nil {
		t.Fatalf("Could not add new tag ('%s'), err='%s'", newTag.ToString(), err)
	}

	t.Logf("Returned tagId=%d", tagId)

	allTags, err := jcapi.GetAllTags()
	if err != nil {
		t.Fatalf("Could not GetAllTags, err='%s'", err)
	}

	var foundTag int

	for i, tag := range allTags {
		t.Logf("Tag[%d]='%s'", i, tag)
		if tag.Id == tagId {
			foundTag = i
		}
	}

	oneTag, err := jcapi.GetTagByName(allTags[foundTag].Name)
	if err != nil {
		t.Fatalf("Could not get tag by name, '%s', err='%s'", allTags[foundTag].Name, err)
	}
	if oneTag.Name != allTags[foundTag].Name {
		t.Fatalf("Tag names don't match, '%s' != '%s'", oneTag.Name, allTags[foundTag].Name)
	}

	allTags[foundTag].Name = "Test tag 1 with a name change"

	newTagId, err := jcapi.AddUpdateTag(Update, allTags[foundTag])
	if err != nil {
		t.Fatalf("Could not change the test tag's name, err='%s'", err)
	}

	if tagId != newTagId {
		t.Fatalf("The ID of the tag changed during an update, this shouldn't happen.")
	}

	err = jcapi.DeleteTag(allTags[foundTag])
	if err != nil {
		t.Fatalf("Could not delete the tag I just added ('%s'), err='%s'", allTags[foundTag].ToString(), err)
	}
}

func MakeIDSource() JCIDSource {

	return JCIDSource{
		Name:           "Test Name",
		Type:           "Active Directory",
		Version:        "1.0.0",
		IpAddress:      "127.0.0.1",
		LastUpdateTime: "2014-10-14 23:34:33",
		DN:             "CN=JumpCloud;CN=Users;DC=jumpcloud;DC=com",
		Active:         true,
	}
}

func TestIDSources(t *testing.T) {

	jcapi := NewJCAPI(testAPIKey, testUrlBase)

	e := MakeIDSource()

	result, err := jcapi.AddUpdateIDSource(Insert, e)
	if err != nil {
		t.Fatalf("Could not post a new ID Source object, err='%s'", err)
	}

	t.Logf("Post to idsources API successful, result=%U", result)

	extSourceList, err := jcapi.GetAllIDSources()
	if err != nil {
		t.Fatalf("Could not list all external sources, err='%s'", err)
	}

	for idx, source := range extSourceList {
		t.Logf("Result %d: '%s'", idx, source.ToString())
	}

	eGet, exists, err := jcapi.GetIDSourceByName(e.Name)
	if err != nil {
		t.Fatalf("Could not get an external source by name '%s', err='%s'", e.Name, err)
	} else if exists && eGet.Name != e.Name {
		t.Fatalf("Received name is different ('%s') than what was sent ('%s')", eGet.Name, e.Name)
	} else if !exists {
		t.Fatalf("Could not find the record we just put in '%c'")
	}

	//
	// If there's more than one test object with this name, let's just
	// loop over and delete them until we find no more of them...
	//
	for exists, err = true, nil; exists; eGet, exists, err = jcapi.GetIDSourceByName(e.Name) {
		if err != nil {
			t.Fatalf("ERROR: getIDSourceByName() on '%s' failed, err='%s'", eGet.ToString(), err)
		}

		err = jcapi.DeleteIDSource(eGet)
		if err != nil {
			t.Fatalf("ERROR: Delete on '%s' failed, err='%s'", eGet.ToString(), err)
		}
	}
}

func checkAuth(t *testing.T, expectedResult bool, username, password, tag string) {
	authjc := NewJCAPI(testAPIKey, authUrlBase)

	userAuth, err := authjc.AuthUser(username, password, tag)
	if err != nil {
		t.Fatalf("Could not authenticate the user '%s' with password '%s' and tag '%s' err='%s'", username, password, tag, err)
	}

	if userAuth != expectedResult {
		t.Fatalf("userAuth=%t, we expected %s for user='%s', pass='%s', tag='%s'", userAuth, expectedResult, username, password, tag)
	}
}

func TestRestAuth(t *testing.T) {
	jcapi := NewJCAPI(testAPIKey, testUrlBase)

	newUser := MakeTestUser()

	userId, err := jcapi.AddUpdateUser(Insert, newUser)
	if err != nil {
		t.Fatalf("Could not add new user ('%s'), err='%s'", newUser.ToString(), err)
	}
	newUser.Id = userId
	defer jcapi.DeleteUser(newUser)

	t.Logf("Returned userId=%s", userId)

	checkAuth(t, true, newUser.UserName, newUser.Password, "")
	checkAuth(t, false, newUser.UserName, newUser.Password, "mytesttag")
	checkAuth(t, false, newUser.UserName, "a0938mbo", "")
	checkAuth(t, false, "2309vnotauser", newUser.Password, "")
	checkAuth(t, false, "", "", "")

	//
	// Now add a tag and put the user in it, and let's try all the tag checking stuff
	//
	newTag := MakeTestTag()

	newTag.SystemUsers = append(newTag.SystemUsers, userId)

	tagId, err := jcapi.AddUpdateTag(Insert, newTag)
	if err != nil {
		t.Fatalf("Could not add new tag ('%s'), err='%s'", newTag.ToString(), err)
	}
	newTag.Id = tagId
	defer jcapi.DeleteTag(newTag)

	t.Logf("Returned tagId=%d", tagId)

	checkAuth(t, true, newUser.UserName, newUser.Password, newTag.Name)
	checkAuth(t, false, newUser.UserName, newUser.Password, "not a real tag")
	checkAuth(t, true, newUser.UserName, newUser.Password, "")
}
