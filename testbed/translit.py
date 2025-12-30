# -*- coding: utf-8 -*-
import csv
import sys
from collections import OrderedDict

vowels = OrderedDict([
('ँ','n'),
('ं','n'),
('ः','a'),
('अ','a'),
('आ','aa'),
('इ','i'),
('ई','ee'),
('उ','u'),
('ऊ','oo'),
('ऋ','ri'),
('ए','e'),
('ऐ','ae'),
('ओ','o'),
('औ','au'),
('ा','a'),
('ि','i'),
('ी','i'),
('ु','u'),
('ू','oo'),
('ृ','ri'),
('े','e'),
('ै','ai'),
('ो','o'),
('ौ','au')
])

consonants = OrderedDict([
('क','k'),
('क़','q'),
('ख','kh'),
('ग','g'),
('घ','gh'),
('ङ','ng'),

('च','ch'),
('छ','chh'),
('ज','j'),
('ज़','z'),
('ज़','z'), #these two are very different, see them in unicode by 'ज़'.encode('utf-8'). You'll see.
('झ','jh'),
('ञ','nj'),

('ट','t'),
('ठ','th'),
('ड','d'),
('ड़','r'),
('ड़','r'), #these two are very different, see them in unicode by 'ड़'.encode('utf-8'). You'll see.
('ढ','dh'),
('ण','n'),

('त','t'),
('थ','th'),
('द','d'),
('ध','dh'),
('न','n'),

('प','p'),
('फ','ph'),
('फ़','f'),
('फ़','f'), #these two फ़ are very different, see them in unicode by 'फ़'.encode('utf-8'). You'll see.
('ब','b'),
('भ','bh'),
('म','m'),

('य','y'),
('र','r'),
('ल','l'),
('व','v'),
('श','sh'),

('ष','sh'),
('स','s'),
('ह','h'),
('क्ष','ksh'),
('त्र','tr'),
('ज्ञ','gy')
	])

# python3 translit.py "अमर उजाला क़हल"
# python3 translit.py "क़हल"
# python3 translit.py "ख्याति"
# python3 translit.py "ख्याति"
# python3 translit.py "गज़ल"
for x in consonants:
  print(x + ":" + str(x.encode('utf-8')))

str1 = ""
#x = "हवा"
x = sys.argv[1]
for y in x.split():
	#print("y:" + y)
	for i in range(len(y)):
		if (i+1<len(y) and y[i+1].strip()==' ़'.strip()):
			c = y[i]+y[i+1]
			p=2
		else:
			c = y[i]
			p=1

		if (c in vowels.keys()):
			str1 = str1 + vowels[c]
		elif (c in consonants.keys()):
			#print("p:" + str(p) + " c:" + c)
			if(i+p<len(y) and y[i+p] in consonants.keys()):
				if ((c=='झ' and i!=0) or (i!=0 and i+p+1<len(y) and y[i+p+1] in vowels.keys())): # add 'a' after 'jh', only if झ appears in the starting of the word
					str1 = str1 + consonants[c]
				else:
					str1 = str1 + consonants[c]+'a'
			else:
				str1 = str1 + consonants[c]
		elif y[i] in ['\n','\t',' ','!',',','।','-',':','\\','_','?'] or c.isalnum():
			str1 = str1 + c.replace('।','.')
		else:
			print("ignored - " + c)
	str1 = str1 + " "

print(str1.strip())
