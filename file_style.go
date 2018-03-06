package main

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"strings"
	"time"

	"github.com/MJKWoolnough/httpdir"
	"github.com/MJKWoolnough/memio"
)

func init() {
	date := time.Unix(1507215534, 0)
	s := "\xC4:}o\xB7\xF9\x7FK\x9F\x82?\xECX\xA7\x9E+md\xF4\x87.k\xB7\xE8\xD6bنE\xFE\xA0t\x94D\x84\"<J\x96c\xF8\xBB\xE4\xF1\xFD\xC8;9)0\xB8q}w\x9Fw>o\xE4w{Ta\x9A5G\x88\x82\xA7\xF1h'\xF6dV\xACz\x94O\xA35#\x8C\x83%\xB8,\xCB\xF2~<\xAD\xE0\xFA㖳\xAD\n\xFBi\xB3\xD9\xC8O\xB8;\xB0\xB3\xB2\xBC\x92\xCF\xEC\x88\xF8\x86\xB0\x87\xE2\x96`\x87\xAB\n\xD1\xFB\xF1\xE8y<\xEDf\n\xB7@'Q@\x82\xB7,\xC1Q\x81\xB8\x98\xF8\x88\xB8\xDA!\xBC\xDD	-A\xF9\xA8\xB8\xA9Y\x83fr=\\5\x8C\x92\xAF\xAB\xC4\xF9L\xB54\x97\xCCQ\xACw\x98T\x8A~B`\xC1!mj\xC8\xE0\xFF\xF0\xBEf\\@*\xF2h\xFET\xF8\xA8\x90E\xDCx\xF0Q\xE5\xDB\xBC\x9EͿ\x99\x00\xF7\xAB\x9C.nB\x959Z1\ng\xA8=\xE4[L\xADHx\xCC{˵\xBA\xCDk\xC3\xF2[\xCB\xF1W\xAF\xC0\x86\xF1=x\xF5\xD5x<\"p\x85\xB8\x98\xD61\xE6\xB1\xC2G\xF7 \xED9\x82\x8A\xADaP\xE2\x97\xD4}\xFC\xF3\xB2\xACO\x9A\x82B\xF6;G\xB0b\x94<~k\xFE\xF8\x90S\xCD%B\xCA\xEE+\xC6+\xC4;\x8E\xA9_7\xE2\x91 \xB0#\xB8Ҕ\\\xD42\xFF\xBBx\xACѷ\xEBZ\x7F\\\xB1SK\xAB\xC2MM\xE0#X\xCA(rZ\xB2\xFE\xAB\nӭ~\xB4\xA7\n\xCB9h1%X!\x96k/\xA5s1zhwA\x90;E\x85\x00\nuJ\x9F	\xB0%e\xE2\xDAwț\xB4l\xC1R_\x84|E\xD8\xFAc|g%IU\xF88\x85U\xA5\xB7\xF9\x86QQ4\xF8Z\x829\xDA\xC7 \x93\xBC.c\x7FR\x88L\xD0X1\"M=\x8A]B\x85\x9Ef+\xF6\xA0\x83(f\xF5\xA9\xFDUN\xDA7\xC9\xFEs\x91|Q\xF0\xEE\xA9\x91\xE9\xCFJ\xD4\xD1R\xBC\x87Od\xFB.+\xBB\nD&\x00\xB6\"\x96\xD3\xF9\xA2\xB9\xBE\xEF1\xDD\x9B]kX\x9B\x8E\x96;i\xCB8\xE8+m\xAD\xBCQ/j\x86\xBD(-\x97\"\xCEO©\xBC\xF4\xF7\x8A	;3\xB7\xDD\x9E[\xB0\xE29\xE7\x9FV\xFF;ܣ	\xB8\xACɡ\xF9^\xBDP\xF0Ş}*\x8D\xDC删\xB5\xF0vl\xF1\x80V\xB1\xC8|\xF5\xB3\x89\x9Fjs\xB9\xFC\xE2\x85\xF1\xEES\x81i\x85dJ\xB9\xEBr\xDB\xC6\xE4\x90g\xA6=(\xF0\xD1Ka\x8B6i\xA5ekb\x8AuN\"P\xA0\xFF\\\x8B\xF2\xCAf\xE3Cӆ`\xB3\xD7\xB7Q\n\x98\x9D\xE72;8k@\x81\x9A/\xD3}o\xF6\x9E;\xFFк\xFCረhb\xA5uW\xCE\xF4Rk\xDB\xF6\xC9$\xB6\xB6\x8CaB\xB0\xBDy\xEA)\n\xD0Q\xF5(\xBD\xB5\x871\xEBY\x865\xAFB驨n\xB5gT\xEC\x86u`\xE8t\x9FM-H\xB5\xE9_$\xAC\xDD.ww\x81\xF8\xE6\xA9/\xA7\xFE!+\xF3\xE8]\xA85h\x9E\xA4,n3Ɖ\xB90\xB9.	vX\xA0\xA2\xA9\xE1)\xC7~\xE0\xB0>\xA7L\x98\xEE؁\x9B\xF1M\xA4A\xF3\x9C	\x99B\xA7S\x99Z\xBE}m\x90\x97@'\xA5\xB5\x87\xF5n\xD2A1=\xF4\x9B\xE1\x9B\xD8[\xBD\xB2\xDA\xE0\xB7\x9C	(\xD0u\xF1\xB6\xAC\xD0\xF6\xE6l\xA7Ytb\xA4ǻ#mf}Qb\xD5ئ:\xEC=uS\xA2\xDE\xD0G\xF43\xDA\xE5\x99G\xF4\xA5\x9B/Ki\xD2q;m\xD3@*\x9Bw2ٛ\xB4\xAD\xF2=\x9B\x9145\xA4\x818\xF2\x85\x92鈸\xC0kH,\x86=\xAE*\x82:\x9E\xC2\xCCl?:\xED\xF8\xD1\xDC\xA5蚰\xE6\xC0\xD1\xF0\xC6*\xEFT\xAC\xFF3\"\xA4o}\xEC\x9Eݢ\xC1\xA9p\xD6\xF1\xB0\x90D\x90\x88;>\xFBb\x87k\xD3\xE4\xBF(l\xBC\xA5\xA8\xCAZkD\xC8/U\xB6Ἤ\xAA\xCA\x95\xF97!쨱\xFD\xEBG[V&weYfֽW~\x8D\xF2앛\xDC\xD2\xE6'\xFA\x89\xD4\xC3o[ٺ\xC5\x7Fs\\\xAA\xA8(\xADW\x9F\xDAj\xE6\x80K\xBC\x89<\xE2\xCD`ʵ]\x8FG\xFD\xE1\xAEu\xDBTa%\xCC0\xF5$k\xC4kوLT;r\x93vQ\xA7\x8ET\xB18I:뢉\xB2\xC3َ\xEBH\xEAh\xD9gu\xE70kH\xA5_\xA2\xB0\xFFP\x91!1И\xB7\xDA\xFA\x9Fv\x82\xD6\xE6\xB8\xC7D\xD6\xE2xӱA(\xD9\xD1\xE69\x94\xA9>O\xB7h\x99>\xEFR\xB0\xFA\xF4\x9B\x96NZ\x88\x9B\xBF\x8C\xDBl3\x9F.\xAER\xD8\xE39\x98\xEB\xB4켬\xC1\xFB\x9A\xA0w!u\xE0\xA7{\xD44p\x8B\xDE#*\xCE	e\xAE{\x8A\x89\x86\xEEE3\xF9l[\xD5Ҡ\xF1\xE7V\xB4\xAF0\xA3K\xA0^r0o\x00\xA6L\xB1\x8D\xFC\xDF}D\x8F\xF7\xA810rey\x9E@N\x99\xE0y<-\xB2 \x92'\"ˍ4\x8As\x8A \xFF\xB5WL/3X1\xC3U\x9E\xA8\xF6À\xB8ܗ\x8A,\xF9{\x91Ⱦ\xF9\x9B\xB0\xBB\xE9r\x81/\x8A\xAF)o\x87\xC1\xD6L\xA8\xB7\xC1\xE8\xC9J)\x8En\x84m,:\xC3P޲\xDDN\xC8\xC1}\x896\xD2\xF5.\xD6RP\xFD\xD5ft\xE4osf\xA1;\xD3\xE4wZ\xFC\x98\xB9%\xBB6]\xCB\xC4t3\x83 `Z\xE1\xEEWx{\x80\x82\xF1\x81\x8A\xCA\xC8$\x83e\xB2\xDA\xEA!Ǫj\x80!	q.?\xE1\xDC<\xE4\xC7s*\x95\x9F\xCFj\xE9E\xF0\xFC~\x88\xF8I\x96@\xAD\xC7hNV\xECt?P\x8AY'[\xE4ͫr\xAD<\xD4'\xF0^g\xDA\xCFo\xAC2\xE9Q\xAB'\xAA\xED\x933\xEF\xA8ʚ\xA2Si\xA5\xCE :I\xC5hpրF\xA0\xBA@\xB4\x9A\xB8\xACM\x8Cjn\xE7x\xF0\xF0yƙ\xE2\xF5 ޲$3\xA5f\xA6\x90\x8B!v\xBC\x82!<r*\x87G'\x85\x9D\xD0s7\xF9\xF6\xF0\xCC\xFAĒ,Mޣԁ\\.\xA0$N\xFE\xFAXr\xF8\xA5[đ\xD6x\xB4\xEE\x88\xBC\xC2\xE9@ꁠ\xFB\xEEql\xDB\xF6u\xB6z\xE2|i\xB8\x00\xCD\xF4\xDBuy\xD3\xD3\xEA$wY\xFFi@\xBC`{݉\xD9o\xB9\xEF9\x89&W\xA6\xF86*/\x84\x85v\xEC\xAA30\xA2\xD3c9=\xB3\xAC+0\xF5'\xE7\x9D\xD8\xDDz@\x8E\xF096\x9B\xD96UX|\x8F$\xEA\xC0\xBF\xF5A\x82\xEB\xD5\xF9\xAE*\xDF\xEDIo6Az8\xBBG\xD9\xF3y\x80B'\xE0\xAA\xA8\x87\x83\xAA\xF9]\xE7\xCC\xC5?\xC9\xE4\x9E\x98\xBB\xAA\xFFHA2\xB1\x84\x9B\x81\"iͨ@*\x85\\\\\xA4\x98\x8F\xFC\xE8f\x99H%\xB0pJ\x94\xE7%\xE7>\xBF]R\xBE\x9B@\xCB\xDEm\x80|*=k\xE8\xED\x9B\"b\xDDs\xAAb\xD1I\xB8\x93\x87V#\x82\xD5mK^pX\xE1C\xE3\xC6$\xEE\xBB*D\xE2\xCF6\xC8H/ò\x90/\xB6Qq-h5=\x97h-\x7F\xA4_N\xC0e5\x93?\xCA^7\xB6\xD80\xCD~	^\xD7'\xF5Oݭ('@\xFF7\xBD\xBB\x99\x00L$\xCC)mf\\\xD0~5\xDF>\xB7\xB0\x88\xEC\xE5\xE5೅\xDE@fR\xE0ͦG\xFA\\\xE3\xD1^\xA1\xE3h\xFCw\x83\xDB\xC4\xE5\xEAYt^\x91\x9E\x88y\xDB\n\\\xF4zi\xFE\xFCq\xF1u\xE0?-/\x81\x8B\x81\xD8\xC7܉G\x945\x8Cs\xA8i\x91\xFAW-&xqz13\xA64/\x9E;\xF7\xB0b\x99\x89Y)\xCE\xE0%\x9A\x9B\xE4\xF3' \xF3\xCE\xC6\xD209\xEC9\xD7\xECt\x8E\xDA\xC0}\xA8<Kf?\xC0\x8BS\x7FN[1\x82[05\xAE\xF7\xB7\x89\xFA>Bxܬ9#$\x8Ch\xDEŅi\xCD1\xBFP\xF2\xD8S\xF2czdx\x8D\xFE\xC9j \xAA\xDF5}\x99Re\xBD\xFD\xAD\xE0\xF4a\xBBa\x90A,\xEF\xDA\xFD\x98\xBCԀ\x7FaL >Q4i\xD8\xFCx\xE6k\xF9c\xD5˶[\x82ޱ\x93\xBE\xF4\xE5\xD4\xEBn\x92姹6^h\xD0J/2E\x9ChtYX\xE0eg!\x85l)\xF8\xD2_\x9A	=\x95UvX\x82\x8B\x7F\xB7\x85\xFF\xC5\xC0\xBC&Q=\xFB\xF7f\xF2\xD28J?\xD1\xE3g\xD32\xF7\x88\x98w\xF7L\xDDZB\xD5\xB8\xEDJ\xEB\x9D\xE3\xFAH\x99\xC6c\xD1Dv\xC9`\xB1f\xE5\xEC\xE1\xFD\xAC\xEB\x94-Yײ\xDB:\xBB\xA87\xD8\xD2i#\xFF/\x83\x84\xC4n\xBA)X\xB8\xE9\xDA-\x82\xA9+\x80_\xAB{o^\xA4\xF3a\x9B\xAA\xE2S{\x9Bs\x93)3\xA4\xB5D\xFBZ<&\xAF\x99\xF9+C\xF8\xF1\xD6\xFE\x9DQ\xF3\xF3x\xAC/\n\xAB\x88\x9E\xC6\xE3ѫn\xA3\xEA*W+T9x\xC1c\xF8\x8E\xD5U\xF7Vp:\xD4&\xAF\xD2\xFA\xE5-{	yv\xF3	ؽ\x9E\x80\xDD\xDD\xC0\x85d\xFFZ^x8\x93\xE6D\xB0_\xA5\xA2T\xB4R\x7F\x81Wq\xE5\xB8q\xDA\xD2\xEF\xA2\xA8U\x9F\xB7\xC4)\xD1{it\xE6\xBDJ\x97\x96\x97\xC2΢R\xFA\xA7*+\xB7\xB9\xE9;\xF1.\x8E\xF6[\xF6\x95\xC1\xC5_u\xD3\xF1B\xE2\xBA\xF8\x90\xE6\xE5y<67\xD4\xC3\xE3\xB4?q%\xF7\xF3xl%\x89G\xF3d\xE0\x94\xED2QM\xB4*\x84\xB2\xB2\xDFϨ\xC7A\xFE&*\xFF\xEEk\xE4\x89I\xEFz\x8F;\xB5I\xEA޴\x91\xE2NX\xBD8\xD9\xC3\x87mm\xBD\xF2<\xDB̜m\xF8\xA2&ޞ\xD6-Z\x9A\x81\x00:\xEA\xBC.\xB8\xB8ز\xBC\xE8O\xB7\xA6\x86\xA8v\xA3\xA8e\x99\xEBPAܜMg2rJR~\xAD\xDE\xC1\xBF\x95\x96\xCD\xE3\xAB\xF5\xED۫.\x86\xAA{s\xDA\xD5XB\xEC\ng\xE6빺[\x9D[%Ч\xA7\xB4l/\xCF\xF2\xE0\xDC\xD5\xC3(\xD19TA\xB5x>\x99_r]\xCC\xC9,ŉ;\x8DH\xF3]?\xB9殻\xA6\xF5H\x7F\xC5\xD3\xD8\xC5E5\xC6*\x83\xEF\xB6>\xD7\xE2\xE9\xB7@T)\xA9\xE7\xF91l}G*\x99ig\x97H\xB0\xDD{\xB5V\xB8\x97\x8A\x8E\xE5\xCCr>\x95Q\xA1\xCB\xDFP\xE5\x91G\xA5w_Ii\xB1C\xE3$\xB2\x8AV\xFD\xDCǝ\xEB\xE6RX;\xE4_\xCBKauB>\x95]\xA0VB\xF7\xBC+\xF6\x7F\x00\x00\xFF\xFF"
	b := make([]byte, 13638)
	fl := flate.NewReader(strings.NewReader(s))
	io.ReadFull(fl, b)
	fl.Close()
	httpdir.Default.Create("style.css.fl", httpdir.FileString(s, date))
	gzb := make(memio.Buffer, 0, 2940)
	gz, _ := gzip.NewWriterLevel(&gzb, gzip.BestCompression)
	gz.Write(b)
	gz.Close()
	httpdir.Default.Create("style.css.gz", httpdir.FileBytes(gzb, date))
	httpdir.Default.Create("style.css", httpdir.FileBytes(b, date))
}
