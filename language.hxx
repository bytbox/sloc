#ifndef LANGUAGE_HXX
#define LANGUAGE_HXX

#include <map>
#include <string>

class Language {
public:
	virtual ~Language();
protected:
private:
};

extern std::map<std::string, Language> language_index;

void init_langauges();

// A dumb language which just counts (blank- and non-blank) lines.
class LineLanguage : public Language {
public:
	LineLanguage();
	~LineLanguage();
protected:
private:
};

#endif /* !LANGUAGE_HXX */

