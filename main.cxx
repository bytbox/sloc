#include "language.hxx"

#include <fstream>
#include <iostream>
#include <sstream>
#include <string>
#include <vector>
using namespace std;

#include <dirent.h>
#include <sys/stat.h>
#include <unistd.h>

void recursive_add(string from, vector<string> &to);

int main(int argc, char *argv[]) {
	vector<string> args;
	if (argc > 1) {
		for (int i=1; i<argc; i++)
			args.push_back(argv[i]);
	} else {
		args.push_back(".");
	}

	vector<string> files;
	for (vector<string>::iterator i = args.begin(); i != args.end(); i++) {
		recursive_add(*i, files);
	}

	for (vector<string>::iterator i = files.begin(); i != files.end(); i++) {
	}
	return 0;
}

void recursive_add(string from, vector<string> &to) {
	// find out what 'from' is
	struct stat st;
	if (stat(from.c_str(), &st)) {
		cerr << "  ! " << from << endl;
		return;
	}
	if (S_ISREG(st.st_mode)) {
		to.push_back(from);
		return;
	}
	if (S_ISDIR(st.st_mode)) {
		// read the directory and recursive_add all entries
		DIR *d = opendir(from.c_str());
		dirent *de;
		while ((de = readdir(d))) {
			if (de->d_name[0] != '.') {
				ostringstream oss;
				oss << from << "/" << de->d_name;
				recursive_add(oss.str(), to);
			}
		}
		closedir(d);
		return;
	}
	// ignore it
}

