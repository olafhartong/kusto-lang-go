package kustolang

import "testing"

func TestConvertFunctionsCatalog(t *testing.T) {
	if len(Functions.All) != 429 {
		t.Fatalf("function count = %d, want 429", len(Functions.All))
	}
	if Functions.All[0] != Functions.Cluster ||
		Functions.All[1] != Functions.Database ||
		Functions.All[2] != Functions.Table ||
		Functions.All[3] != Functions.ExternalTable ||
		Functions.All[4] != Functions.MaterializedView ||
		Functions.All[5] != Functions.EntityGroup ||
		Functions.All[6] != Functions.StoredQueryResult ||
		Functions.All[7] != Functions.Graph ||
		Functions.All[8] != Functions.Strcat ||
		Functions.All[9] != Functions.StrcatArray ||
		Functions.All[10] != Functions.ArrayStrcat ||
		Functions.All[11] != Functions.StrcatDelim ||
		Functions.All[12] != Functions.Strcmp ||
		Functions.All[13] != Functions.Strrep ||
		Functions.All[14] != Functions.Strlen ||
		Functions.All[15] != Functions.StringSize ||
		Functions.All[16] != Functions.ToUpper ||
		Functions.All[17] != Functions.ToLower ||
		Functions.All[18] != Functions.ToUtf8Deprecated ||
		Functions.All[19] != Functions.UnicodeCodepointsFromString ||
		Functions.All[20] != Functions.Substring ||
		Functions.All[21] != Functions.RegexQuote ||
		Functions.All[22] != Functions.IndexOf ||
		Functions.All[23] != Functions.IndexOfRegex ||
		Functions.All[24] != Functions.HasAnyIndex ||
		Functions.All[25] != Functions.Reverse ||
		Functions.All[26] != Functions.Split ||
		Functions.All[27] != Functions.ParseCommandLine ||
		Functions.All[28] != Functions.Extract ||
		Functions.All[29] != Functions.ExtractAllDeprecated ||
		Functions.All[30] != Functions.ExtractAll ||
		Functions.All[31] != Functions.ExtractJsonDeprecated ||
		Functions.All[32] != Functions.ExtractJson ||
		Functions.All[33] != Functions.Replace ||
		Functions.All[34] != Functions.ReplaceRegex ||
		Functions.All[35] != Functions.ReplaceString ||
		Functions.All[36] != Functions.ReplaceStrings ||
		Functions.All[37] != Functions.TrimStart ||
		Functions.All[38] != Functions.TrimEnd ||
		Functions.All[39] != Functions.Trim ||
		Functions.All[40] != Functions.CountOf ||
		Functions.All[41] != Functions.Translate ||
		Functions.All[42] != Functions.MakeStringDeprecated ||
		Functions.All[43] != Functions.UnicodeCodepointsToString ||
		Functions.All[44] != Functions.DatetimeLocalToUTC ||
		Functions.All[45] != Functions.DatetimeUTCToLocal ||
		Functions.All[46] != Functions.DateTimeListTimezones ||
		Functions.All[47] != Functions.ToString ||
		Functions.All[62] != Functions.GetType ||
		Functions.All[63] != Functions.UrlEncode ||
		Functions.All[78] != Functions.Lz4CompressDynamicArray ||
		Functions.All[79] != Functions.ParseCsv ||
		Functions.All[93] != Functions.ParseVersion ||
		Functions.All[94] != Functions.FormatDatetime ||
		Functions.All[124] != Functions.UnixTimeNanosecondsToDateTime ||
		Functions.All[125] != Functions.HashCrc32 ||
		Functions.All[135] != Functions.HashManyCrc32 ||
		Functions.All[136] != Functions.Iif ||
		Functions.All[139] != Functions.Assert ||
		Functions.All[140] != Functions.Bin ||
		Functions.All[143] != Functions.BinAuto ||
		Functions.All[144] != Functions.Not ||
		Functions.All[156] != Functions.Around ||
		Functions.All[157] != Functions.BinaryAnd ||
		Functions.All[163] != Functions.BitsetCountOnes ||
		Functions.All[164] != Functions.TreePath ||
		Functions.All[179] != Functions.ArrayRotateRight ||
		Functions.All[180] != Functions.ArraySortAsc ||
		Functions.All[183] != Functions.Zip ||
		Functions.All[184] != Functions.Pack ||
		Functions.All[185] != Functions.PackDictionary ||
		Functions.All[186] != Functions.BagPack ||
		Functions.All[187] != Functions.BagPackColumns ||
		Functions.All[188] != Functions.PackAll ||
		Functions.All[189] != Functions.PackArray ||
		Functions.All[190] != Functions.SetHasElement ||
		Functions.All[191] != Functions.SetUnion ||
		Functions.All[192] != Functions.SetIntersect ||
		Functions.All[193] != Functions.SetDifference ||
		Functions.All[194] != Functions.SetEquals ||
		Functions.All[195] != Functions.BagMerge ||
		Functions.All[196] != Functions.DynamicToJson ||
		Functions.All[197] != Functions.BagRemoveKeys ||
		Functions.All[198] != Functions.BagHasKey ||
		Functions.All[199] != Functions.JaccardIndex ||
		Functions.All[200] != Functions.BagSetKey ||
		Functions.All[201] != Functions.BagZip ||
		Functions.All[202] != Functions.PunycodeDecode ||
		Functions.All[203] != Functions.PunycodeEncode ||
		Functions.All[204] != Functions.PunycodeDomainDecode ||
		Functions.All[205] != Functions.PunycodeDomainEncode ||
		Functions.All[206] != Functions.PercentileTDigest ||
		Functions.All[207] != Functions.PercentileArrayTDigest ||
		Functions.All[208] != Functions.PercentRankTDigest ||
		Functions.All[209] != Functions.RankTDigest ||
		Functions.All[210] != Functions.TdigestIsValid ||
		Functions.All[211] != Functions.HllIsValid ||
		Functions.All[212] != Functions.TDigestMerge ||
		Functions.All[213] != Functions.MergeTDigest ||
		Functions.All[214] != Functions.HllMerge ||
		Functions.All[215] != Functions.DCountHll ||
		Functions.All[216] != Functions.HllNormalize ||
		Functions.All[217] != Functions.SeriesFir ||
		Functions.All[218] != Functions.SeriesStats ||
		Functions.All[219] != Functions.SeriesStatsDynamic ||
		Functions.All[220] != Functions.SeriesFft ||
		Functions.All[221] != Functions.SeriesIfft ||
		Functions.All[222] != Functions.SeriesFitLine ||
		Functions.All[223] != Functions.SeriesFitLineDynamic ||
		Functions.All[224] != Functions.SeriesFit2Lines ||
		Functions.All[225] != Functions.SeriesFit2LinesDynamic ||
		Functions.All[226] != Functions.SeriesOutliers ||
		Functions.All[227] != Functions.SeriesIIR ||
		Functions.All[228] != Functions.SeriesPeriodsDetect ||
		Functions.All[229] != Functions.SeriesPeriodsValidate ||
		Functions.All[230] != Functions.SeriesFillBackwards ||
		Functions.All[231] != Functions.SeriesFillForward ||
		Functions.All[232] != Functions.SeriesFillConst ||
		Functions.All[233] != Functions.SeriesFillLinear ||
		Functions.All[234] != Functions.SeriesFitPoly ||
		Functions.All[235] != Functions.SeriesAdd ||
		Functions.All[236] != Functions.SeriesSubtract ||
		Functions.All[237] != Functions.SeriesMultiply ||
		Functions.All[238] != Functions.SeriesDivide ||
		Functions.All[239] != Functions.SeriesPow ||
		Functions.All[240] != Functions.SeriesGreater ||
		Functions.All[241] != Functions.SeriesGreaterEquals ||
		Functions.All[242] != Functions.SeriesLess ||
		Functions.All[243] != Functions.SeriesLessEquals ||
		Functions.All[244] != Functions.SeriesEquals ||
		Functions.All[245] != Functions.SeriesNotEquals ||
		Functions.All[246] != Functions.SeriesExp ||
		Functions.All[247] != Functions.SeriesSign ||
		Functions.All[248] != Functions.SeriesAbs ||
		Functions.All[249] != Functions.SeriesSin ||
		Functions.All[250] != Functions.SeriesAsin ||
		Functions.All[251] != Functions.SeriesCos ||
		Functions.All[252] != Functions.SeriesAcos ||
		Functions.All[253] != Functions.SeriesTan ||
		Functions.All[254] != Functions.SeriesAtan ||
		Functions.All[255] != Functions.SeriesMagnitude ||
		Functions.All[256] != Functions.SeriesSum ||
		Functions.All[257] != Functions.SeriesProduct ||
		Functions.All[258] != Functions.SeriesLog ||
		Functions.All[259] != Functions.SeriesFloor ||
		Functions.All[260] != Functions.SeriesCeiling ||
		Functions.All[261] != Functions.ArraySum ||
		Functions.All[262] != Functions.SeriesSeasonal ||
		Functions.All[263] != Functions.SeriesDecompose ||
		Functions.All[264] != Functions.SeriesDecomposeForecast ||
		Functions.All[265] != Functions.SeriesDecomposeAnomalies ||
		Functions.All[266] != Functions.SeriesPearsonCorrelation ||
		Functions.All[267] != Functions.SeriesDotProduct ||
		Functions.All[268] != Functions.SeriesCosineSimilarity ||
		Functions.All[269] != Functions.Round ||
		Functions.All[270] != Functions.Ceiling ||
		Functions.All[271] != Functions.Pow ||
		Functions.All[272] != Functions.Sqrt ||
		Functions.All[273] != Functions.Log ||
		Functions.All[274] != Functions.Log2 ||
		Functions.All[275] != Functions.Log10 ||
		Functions.All[276] != Functions.Exp ||
		Functions.All[277] != Functions.Exp2 ||
		Functions.All[278] != Functions.Exp10 ||
		Functions.All[279] != Functions.PI ||
		Functions.All[280] != Functions.Cos ||
		Functions.All[281] != Functions.Sin ||
		Functions.All[282] != Functions.Tan ||
		Functions.All[283] != Functions.Acos ||
		Functions.All[284] != Functions.Asin ||
		Functions.All[285] != Functions.Atan ||
		Functions.All[286] != Functions.Atan2 ||
		Functions.All[287] != Functions.Abs ||
		Functions.All[288] != Functions.Cot ||
		Functions.All[289] != Functions.Degrees ||
		Functions.All[290] != Functions.Radians ||
		Functions.All[291] != Functions.Sign ||
		Functions.All[292] != Functions.Rand ||
		Functions.All[293] != Functions.BetaCdf ||
		Functions.All[294] != Functions.BetaInv ||
		Functions.All[295] != Functions.BetaPdf ||
		Functions.All[296] != Functions.Gamma ||
		Functions.All[297] != Functions.LogGamma ||
		Functions.All[298] != Functions.Erf ||
		Functions.All[299] != Functions.Erfc ||
		Functions.All[300] != Functions.IsNan ||
		Functions.All[301] != Functions.IsInf ||
		Functions.All[302] != Functions.IsFinite ||
		Functions.All[303] != Functions.Coalesce ||
		Functions.All[304] != Functions.MaxOf ||
		Functions.All[305] != Functions.MinOf ||
		Functions.All[306] != Functions.WelchTest ||
		Functions.All[307] != Functions.GeoFromWkt ||
		Functions.All[308] != Functions.GeoAngle ||
		Functions.All[309] != Functions.GeoAzimuth ||
		Functions.All[310] != Functions.GeoClosestPointOnLine ||
		Functions.All[311] != Functions.GeoClosestPointOnPolygon ||
		Functions.All[312] != Functions.GeoDistance2Points ||
		Functions.All[313] != Functions.GeoDistancePointToLine ||
		Functions.All[314] != Functions.GeoDistancePointToPolygon ||
		Functions.All[315] != Functions.GeoPointInCircle ||
		Functions.All[316] != Functions.GeoPointInPolygon ||
		Functions.All[317] != Functions.GeoIntersects2Lines ||
		Functions.All[318] != Functions.GeoIntersectsLineWithPolygon ||
		Functions.All[319] != Functions.GeoIntersects2Polygons ||
		Functions.All[320] != Functions.GeoIntersection2Lines ||
		Functions.All[321] != Functions.GeoIntersectionLineWithPolygon ||
		Functions.All[322] != Functions.GeoIntersection2Polygons ||
		Functions.All[323] != Functions.GeoPolygonsUnion ||
		Functions.All[324] != Functions.GeoSimplifyPolygonsArray ||
		Functions.All[325] != Functions.GeoLinesUnion ||
		Functions.All[326] != Functions.GeoPolygonToH3Cells ||
		Functions.All[327] != Functions.GeoPolygonToS2Cells ||
		Functions.All[328] != Functions.GeoPolygonS2CellCoveringLevel ||
		Functions.All[329] != Functions.GeoPolygonDensify ||
		Functions.All[330] != Functions.GeoPolygonArea ||
		Functions.All[331] != Functions.GeoPolygonBuffer ||
		Functions.All[332] != Functions.GeoPolygonCentroid ||
		Functions.All[333] != Functions.GeoPolygonValidate ||
		Functions.All[334] != Functions.GeoPolygonPerimeter ||
		Functions.All[335] != Functions.GeoPolygonSimplify ||
		Functions.All[336] != Functions.GeoLineS2CellCoveringLevel ||
		Functions.All[337] != Functions.GeoLineLength ||
		Functions.All[338] != Functions.GeoLineBuffer ||
		Functions.All[339] != Functions.GeoLineCentroid ||
		Functions.All[340] != Functions.GeoLineDensify ||
		Functions.All[341] != Functions.GeoLineSimplify ||
		Functions.All[342] != Functions.GeoLineLocatePoint ||
		Functions.All[343] != Functions.GeoLineInterpolatePoint ||
		Functions.All[344] != Functions.GeoLineValidate ||
		Functions.All[345] != Functions.GeoLineToS2Cells ||
		Functions.All[346] != Functions.GeoLengthToS2CellLevel ||
		Functions.All[347] != Functions.GeoPointToGeohash ||
		Functions.All[348] != Functions.GeohashToCentralPoint ||
		Functions.All[349] != Functions.GeohashToPolygon ||
		Functions.All[350] != Functions.GeohashNeighbors ||
		Functions.All[351] != Functions.GeoPointBuffer ||
		Functions.All[352] != Functions.GeoPointToS2Cell ||
		Functions.All[353] != Functions.GeoS2CellToCentralPoint ||
		Functions.All[354] != Functions.GeoS2CellToPolygon ||
		Functions.All[355] != Functions.GeoS2CellNeighbors ||
		Functions.All[356] != Functions.GeoPointToH3Cell ||
		Functions.All[357] != Functions.GeoH3CellToCentralPoint ||
		Functions.All[358] != Functions.GeoH3CellToPolygon ||
		Functions.All[359] != Functions.GeoH3CellNeighbors ||
		Functions.All[360] != Functions.GeoH3CellChildren ||
		Functions.All[361] != Functions.GeoH3CellParent ||
		Functions.All[362] != Functions.GeoH3CellRings ||
		Functions.All[363] != Functions.GeoH3CellLevel ||
		Functions.All[364] != Functions.GraphAll ||
		Functions.All[365] != Functions.Any ||
		Functions.All[366] != Functions.Map ||
		Functions.All[367] != Functions.InnerNodes ||
		Functions.All[368] != Functions.NodeId ||
		Functions.All[369] != Functions.NodeDegreeIn ||
		Functions.All[370] != Functions.NodeDegreeOut ||
		Functions.All[371] != Functions.Labels ||
		Functions.All[372] != Functions.Ipv4Compare ||
		Functions.All[373] != Functions.Ipv4IsMatch ||
		Functions.All[374] != Functions.Ipv6Compare ||
		Functions.All[375] != Functions.Ipv4IsPrivate ||
		Functions.All[376] != Functions.Ipv6IsMatch ||
		Functions.All[377] != Functions.Ipv6IsInRange ||
		Functions.All[378] != Functions.Ipv6IsInAnyRange ||
		Functions.All[379] != Functions.Ipv4IsInRange ||
		Functions.All[380] != Functions.Ipv4IsInAnyRange ||
		Functions.All[381] != Functions.Ipv4NetmaskSuffix ||
		Functions.All[382] != Functions.Ipv6LookupRanges ||
		Functions.All[383] != Functions.FormatIPV4 ||
		Functions.All[384] != Functions.FormatIPV4Mask ||
		Functions.All[385] != Functions.FormatBytes ||
		Functions.All[386] != Functions.ConvertAngle ||
		Functions.All[393] != Functions.ConvertVolume ||
		Functions.All[394] != Functions.CurrentClusterEndpoint ||
		Functions.All[395] != Functions.CurrentDatabase ||
		Functions.All[396] != Functions.CurrentPrincipal ||
		Functions.All[397] != Functions.CurrentPrincipalDetails ||
		Functions.All[398] != Functions.CurrentPrincipalIsMemberOf ||
		Functions.All[399] != Functions.ExtentId ||
		Functions.All[400] != Functions.ExtentId2 ||
		Functions.All[401] != Functions.ExtentTags ||
		Functions.All[402] != Functions.CurrentNodeId ||
		Functions.All[403] != Functions.IngestionTime ||
		Functions.All[404] != Functions.RowId ||
		Functions.All[405] != Functions.CursorAfter ||
		Functions.All[406] != Functions.CursorBeforeOrAt ||
		Functions.All[407] != Functions.CursorCurrent ||
		Functions.All[408] != Functions.CursorCurrent2 ||
		Functions.All[409] != Functions.HasIpv4 ||
		Functions.All[410] != Functions.HasIpv4Prefix ||
		Functions.All[411] != Functions.HasAnyIpv4 ||
		Functions.All[412] != Functions.HasAnyIpv4Prefix ||
		Functions.All[413] != Functions.Ipv4RangeToCidrList ||
		Functions.All[414] != Functions.RowNumber ||
		Functions.All[415] != Functions.RowCumSum ||
		Functions.All[416] != Functions.RowRank ||
		Functions.All[417] != Functions.RowRankDense ||
		Functions.All[418] != Functions.RowRankMin ||
		Functions.All[419] != Functions.RowWindowSession ||
		Functions.All[420] != Functions.Prev ||
		Functions.All[421] != Functions.Next ||
		Functions.All[422] != Functions.RowstoreOrdinalRange ||
		Functions.All[423] != Functions.EstimateDataSize ||
		Functions.All[424] != Functions.NewGuid ||
		Functions.All[425] != Functions.Invoke ||
		Functions.All[426] != Functions.Cast ||
		Functions.All[427] != Functions.IpGeoLocation ||
		Functions.All[428] != Functions.ColumnNamesOf {
		t.Fatalf("function All ordering = %#v", Functions.All)
	}

	checkSourceReferenceFunctions(t)
	checkStringFunctions(t)
	checkTypeConversionFunction(t, Functions.ToString, "tostring", ScalarTypes.String.Type(), ParameterTypeScalar, ResultNameFirstArgument)
	checkTypeConversionFunction(t, Functions.ToHex, "tohex", ScalarTypes.String.Type(), ParameterTypeInteger, ResultNameFirstArgument)
	checkTypeConversionFunction(t, Functions.ToDynamic, "todynamic", ScalarTypes.Dynamic.Type(), ParameterTypeDeclared, ResultNameFirstArgument)
	checkTypeConversionFunction(t, Functions.ToLong, "tolong", ScalarTypes.Long.Type(), ParameterTypeScalar, ResultNameFirstArgument)
	checkTypeConversionFunction(t, Functions.ToInt, "toint", ScalarTypes.Int.Type(), ParameterTypeScalar, ResultNameFirstArgument)
	checkTypeConversionFunction(t, Functions.ToReal, "toreal", ScalarTypes.Real.Type(), ParameterTypeScalar, ResultNameFirstArgument)
	checkTypeConversionFunction(t, Functions.ToDouble, "todouble", ScalarTypes.Real.Type(), ParameterTypeScalar, ResultNameFirstArgument)
	checkTypeConversionFunction(t, Functions.ToDateTime, "todatetime", ScalarTypes.DateTime.Type(), ParameterTypeScalar, ResultNameFirstArgument)
	checkTypeConversionFunction(t, Functions.ToTimespan, "totimespan", ScalarTypes.TimeSpan.Type(), ParameterTypeScalar, ResultNameFirstArgument)
	checkTypeConversionFunction(t, Functions.ToBool, "tobool", ScalarTypes.Bool.Type(), ParameterTypeScalar, ResultNameFirstArgument)
	checkTypeConversionFunction(t, Functions.ToBoolean, "toboolean", ScalarTypes.Bool.Type(), ParameterTypeScalar, ResultNameFirstArgument)
	checkTypeConversionFunction(t, Functions.ToDecimal, "todecimal", ScalarTypes.Decimal.Type(), ParameterTypeScalar, ResultNameFirstArgument)
	checkTypeConversionFunction(t, Functions.ToGuid, "toguid", ScalarTypes.Guid.Type(), ParameterTypeStringOrDynamic, ResultNameFirstArgument)
	if !Functions.ToObjectDeprecated.IsHidden() || Functions.ToObjectDeprecated.Alternative() != "todynamic" {
		t.Fatalf("toobject obsolete metadata = hidden:%t alt:%q", Functions.ToObjectDeprecated.IsHidden(), Functions.ToObjectDeprecated.Alternative())
	}
	if !Functions.ToTime.IsHidden() || Functions.ToTime.Alternative() != "totimespan" {
		t.Fatalf("totime obsolete metadata = hidden:%t alt:%q", Functions.ToTime.IsHidden(), Functions.ToTime.Alternative())
	}
	if Functions.GetType.ResultNameKind() != ResultNamePrefixAndFirstArgument || Functions.GetType.ResultNamePrefix() != "type" {
		t.Fatalf("gettype result metadata = %s %q", Functions.GetType.ResultNameKind(), Functions.GetType.ResultNamePrefix())
	}

	checkEncodingFunction(t, Functions.UrlEncode, "url_encode", ScalarTypes.String.Type(), "url", ParameterTypeDeclared)
	checkEncodingFunction(t, Functions.UrlEncodeComponent, "url_encode_component", ScalarTypes.String.Type(), "url", ParameterTypeDeclared)
	checkEncodingFunction(t, Functions.UrlDecode, "url_decode", ScalarTypes.String.Type(), "encoded_url", ParameterTypeDeclared)
	checkEncodingFunction(t, Functions.Base64EncodeToString, "base64_encode_tostring", ScalarTypes.String.Type(), "string", ParameterTypeDeclared)
	checkEncodingFunction(t, Functions.Base64DecodeToArray, "base64_decode_toarray", ScalarTypes.DynamicArray.Type(), "base64_string", ParameterTypeDeclared)
	checkEncodingFunction(t, Functions.Base64EncodeFromArray, "base64_encode_fromarray", ScalarTypes.String.Type(), "base64_string_decoded_as_array", ParameterTypeDynamicArray)
	checkEncodingFunction(t, Functions.Base64DecodeToGuid, "base64_decode_toguid", ScalarTypes.Guid.Type(), "base64_string", ParameterTypeDeclared)
	checkEncodingFunction(t, Functions.Base64EncodeFromGuid, "base64_encode_fromguid", ScalarTypes.String.Type(), "guid", ParameterTypeDeclared)
	checkEncodingFunction(t, Functions.ZlibDecompressString, "zlib_decompress_from_base64_string", ScalarTypes.String.Type(), "string", ParameterTypeDeclared)
	checkEncodingFunction(t, Functions.ZlibCompressString, "zlib_compress_to_base64_string", ScalarTypes.String.Type(), "string", ParameterTypeDeclared)
	checkEncodingFunction(t, Functions.GzipDecompressString, "gzip_decompress_from_base64_string", ScalarTypes.String.Type(), "string", ParameterTypeDeclared)
	checkEncodingFunction(t, Functions.GzipCompressString, "gzip_compress_to_base64_string", ScalarTypes.String.Type(), "string", ParameterTypeDeclared)
	if !Functions.Base64EncodeString.IsHidden() || Functions.Base64EncodeString.Alternative() != "base64_encode_tostring" {
		t.Fatalf("base64_encodestring obsolete metadata = hidden:%t alt:%q", Functions.Base64EncodeString.IsHidden(), Functions.Base64EncodeString.Alternative())
	}
	if !Functions.Base64DecodeString.IsHidden() || Functions.Base64DecodeString.Alternative() != "base64_decode_tostring" {
		t.Fatalf("base64_decodestring obsolete metadata = hidden:%t alt:%q", Functions.Base64DecodeString.IsHidden(), Functions.Base64DecodeString.Alternative())
	}
	if !Functions.Lz4CompressDynamicArray.IsHidden() || Functions.Lz4CompressDynamicArray.ResultNameKind() != ResultNameNone {
		t.Fatalf("lz4 helper metadata = hidden:%t result:%s", Functions.Lz4CompressDynamicArray.IsHidden(), Functions.Lz4CompressDynamicArray.ResultNameKind())
	}

	checkParseFunction(t, Functions.ParseCsv, "parse_csv", ScalarTypes.DynamicArray.Type(), "csv_text", ParameterTypeStringOrDynamic, ResultNameFirstArgument)
	checkParseFunction(t, Functions.ParseJson, "parse_json", ScalarTypes.Dynamic.Type(), "json_text", ParameterTypeStringOrDynamic, ResultNameFirstArgument)
	checkParseFunction(t, Functions.ParseXml, "parse_xml", ScalarTypes.DynamicBag.Type(), "xml_text", ParameterTypeStringOrDynamic, ResultNameFirstArgument)
	checkParseFunction(t, Functions.ParseIPV4, "parse_ipv4", ScalarTypes.Long.Type(), "ip", ParameterTypeDeclared, ResultNameNone)
	checkParseFunction(t, Functions.ParseIPV6, "parse_ipv6", ScalarTypes.String.Type(), "ip", ParameterTypeDeclared, ResultNameNone)
	checkParseFunction(t, Functions.ParseVersion, "parse_version", ScalarTypes.Decimal.Type(), "version", ParameterTypeDeclared, ResultNameNone)
	if !Functions.ParseJsonDeprecated.IsHidden() || Functions.ParseJsonDeprecated.Alternative() != "parse_json" {
		t.Fatalf("parsejson obsolete metadata = hidden:%t alt:%q", Functions.ParseJsonDeprecated.IsHidden(), Functions.ParseJsonDeprecated.Alternative())
	}
	if !Functions.ParseUrlDeprecated.IsHidden() || Functions.ParseUrlDeprecated.Alternative() != "parse_url" {
		t.Fatalf("parseurl obsolete metadata = hidden:%t alt:%q", Functions.ParseUrlDeprecated.IsHidden(), Functions.ParseUrlDeprecated.Alternative())
	}
	if !Functions.ParseUrlQueryDeprecated.IsHidden() || Functions.ParseUrlQueryDeprecated.Alternative() != "parse_urlquery" {
		t.Fatalf("parseurlquery obsolete metadata = hidden:%t alt:%q", Functions.ParseUrlQueryDeprecated.IsHidden(), Functions.ParseUrlQueryDeprecated.Alternative())
	}
	assertDynamicBagProperties(t, Functions.ParseUrl.Signatures()[0].DeclaredReturnSymbol, []string{"Scheme", "Host", "Port", "Path", "Username", "Password", "Query Parameters", "Fragment"})
	assertDynamicBagProperties(t, Functions.ParseUrlQuery.Signatures()[0].DeclaredReturnSymbol, []string{"Query Parameters"})
	assertDynamicBagProperties(t, Functions.ParsePath.Signatures()[0].DeclaredReturnSymbol, []string{"Scheme", "RootPath", "DirectoryPath", "DirectoryName", "Filename", "Extension", "AlternateDataStreamName"})
	assertDynamicBagProperties(t, Functions.ParseUserAgent.Signatures()[0].DeclaredReturnSymbol, []string{"Browser", "OperatingSystem", "Device"})
	userAgentSignature := Functions.ParseUserAgent.Signatures()[0]
	if userAgentSignature.Parameters[1].Name != "look_for" || userAgentSignature.Parameters[1].TypeKind != ParameterTypeStringOrArray || !userAgentSignature.Parameters[1].IsOptional() {
		t.Fatalf("parse_user_agent look_for parameter = %#v", userAgentSignature.Parameters[1])
	}

	checkDateTimeFunction(t, Functions.FormatDatetime, "format_datetime", ScalarTypes.String.Type(), "date", ResultNameFirstArgument)
	if formatParameter := Functions.FormatDatetime.Signatures()[0].Parameters[1]; formatParameter.Name != "format" || formatParameter.ArgumentKind != ArgumentKindLiteralNotEmpty {
		t.Fatalf("format_datetime format parameter = %#v", formatParameter)
	}
	checkDateTimeFunction(t, Functions.FormatTimespan, "format_timespan", ScalarTypes.String.Type(), "timespan", ResultNameFirstArgument)
	checkMakeDatetime(t)
	checkMakeTimespan(t)
	checkDateTimePartFunction(t, Functions.DatetimeAdd, "datetime_add", ScalarTypes.DateTime.Type(), KustoFacts.DateDiffParts, 3)
	checkDateTimePartFunction(t, Functions.DatetimeDiff, "datetime_diff", ScalarTypes.Long.Type(), KustoFacts.DateDiffParts, 3)
	checkDateTimeFunction(t, Functions.DayOfWeek, "dayofweek", ScalarTypes.TimeSpan.Type(), "date", ResultNameNone)
	checkDateTimeFunction(t, Functions.DayOfMonth, "dayofmonth", ScalarTypes.Int.Type(), "date", ResultNameNone)
	checkDateTimeFunction(t, Functions.DayOfYear, "dayofyear", ScalarTypes.Int.Type(), "date", ResultNameNone)
	checkDateTimeFunction(t, Functions.HourOfDay, "hourofday", ScalarTypes.Int.Type(), "date", ResultNameNone)
	if !Functions.WeekOfYear.IsHidden() || Functions.WeekOfYear.Alternative() != "week_of_year" {
		t.Fatalf("weekofyear obsolete metadata = hidden:%t alt:%q", Functions.WeekOfYear.IsHidden(), Functions.WeekOfYear.Alternative())
	}
	checkDateTimeFunction(t, Functions.WeekOfYearISO, "week_of_year", ScalarTypes.Int.Type(), "date", ResultNameNone)
	checkDateTimeFunction(t, Functions.MonthOfYear, "monthofyear", ScalarTypes.Int.Type(), "date", ResultNameNone)
	checkDateTimeFunction(t, Functions.StartOfDay, "startofday", ScalarTypes.DateTime.Type(), "date", ResultNameFirstArgument)
	checkOptionalOffset(t, Functions.StartOfDay, ScalarTypes.Long.Type())
	checkOptionalOffset(t, Functions.StartOfWeek, ScalarTypes.Long.Type())
	checkDateTimeFunction(t, Functions.EndOfYear, "endofyear", ScalarTypes.DateTime.Type(), "date", ResultNameNone)
	checkOptionalOffset(t, Functions.EndOfYear, ScalarTypes.Long.Type())
	checkDateTimeFunction(t, Functions.GetYear, "getyear", ScalarTypes.Int.Type(), "date", ResultNameNone)
	checkDateTimeFunction(t, Functions.GetMonth, "getmonth", ScalarTypes.Int.Type(), "date", ResultNameNone)
	if !Functions.DatePart.IsHidden() || Functions.DatePart.Alternative() != "datetime_part" {
		t.Fatalf("datepart obsolete metadata = hidden:%t alt:%q", Functions.DatePart.IsHidden(), Functions.DatePart.Alternative())
	}
	checkDateTimePartFunction(t, Functions.DatetimePart, "datetime_part", ScalarTypes.Int.Type(), KustoFacts.DateTimeParts, 2)
	checkDateTimeFunction(t, Functions.Now, "now", ScalarTypes.DateTime.Type(), "offset", ResultNameNone)
	if parameter := Functions.Now.Signatures()[0].Parameters[0]; !parameter.IsOptional() || parameter.DeclaredTypes[0] != ScalarTypes.TimeSpan.Type() {
		t.Fatalf("now offset parameter = %#v", parameter)
	}
	checkDateTimeFunction(t, Functions.Ago, "ago", ScalarTypes.DateTime.Type(), "timespan", ResultNameNone)
	if examples := Functions.Ago.Signatures()[0].Parameters[0].Examples; len(examples) != 3 || examples[0] != "30min" || examples[2] != "1d" {
		t.Fatalf("ago examples = %#v", examples)
	}
	checkDateTimeFunction(t, Functions.UnixTimeSecondsToDateTime, "unixtime_seconds_todatetime", ScalarTypes.DateTime.Type(), "number", ResultNameNone)
	checkDateTimeFunction(t, Functions.UnixTimeNanosecondsToDateTime, "unixtime_nanoseconds_todatetime", ScalarTypes.DateTime.Type(), "number", ResultNameNone)
	checkDateTimeFunction(t, Functions.DatetimeLocalToUTC, "datetime_local_to_utc", ScalarTypes.DateTime.Type(), "from", ResultNameNone)
	checkDateTimeFunction(t, Functions.DatetimeUTCToLocal, "datetime_utc_to_local", ScalarTypes.DateTime.Type(), "from", ResultNameNone)
	checkDateTimeFunction(t, Functions.DateTimeListTimezones, "datetime_list_timezones", ScalarTypes.DynamicArray.Type(), "", ResultNameNone)
	checkPunycodeFunctions(t)
	checkDigestSeriesFunctions(t)

	checkHashFunction(t, Functions.HashCrc32, "__hash_crc32", ScalarTypes.Long.Type(), true, 3, 3)
	checkHashFunction(t, Functions.HashDjb2, "__hash_djb2", ScalarTypes.Long.Type(), true, 1, 2)
	checkHashFunction(t, Functions.Hash, "hash", ScalarTypes.Long.Type(), false, 1, 2)
	checkHashFunction(t, Functions.HashSha256, "hash_sha256", ScalarTypes.String.Type(), false, 1, 1)
	checkHashFunction(t, Functions.HashMd5, "hash_md5", ScalarTypes.String.Type(), false, 1, 1)
	checkHashFunction(t, Functions.HashSha1, "hash_sha1", ScalarTypes.String.Type(), false, 1, 1)
	checkHashFunction(t, Functions.HashXXH64, "hash_xxhash64", ScalarTypes.Long.Type(), false, 1, 2)
	checkHashFunction(t, Functions.InternalHashXXH64, "__hash_xxh64", ScalarTypes.Long.Type(), true, 3, 3)
	checkHashFunction(t, Functions.HashCombine, "hash_combine", ScalarTypes.Long.Type(), false, 2, MaxRepeat)
	checkHashFunction(t, Functions.HashMany, "hash_many", ScalarTypes.Long.Type(), false, 1, MaxRepeat)
	checkHashFunction(t, Functions.HashManyCrc32, "__hash_many_crc32", ScalarTypes.Long.Type(), true, 1, 2)

	checkConditionalFunction(t, Functions.Iif, "iif")
	checkConditionalFunction(t, Functions.Iff, "iff")
	checkCaseFunction(t)
	checkAssertFunction(t)
	checkBinFloorFunction(t, Functions.Bin, "bin")
	checkBinFloorFunction(t, Functions.Floor, "floor")
	checkBinAtFunction(t)
	checkBinAutoFunction(t)
	checkBoolStateFunctions(t)
	checkBitwiseFunctions(t)
	checkDynamicArrayFunctions(t)

	checkConvertFunction(t, Functions.ConvertAngle, "convert_angle", 16, "Arcminute", "Tilt")
	checkConvertFunction(t, Functions.ConvertEnergy, "convert_energy", 36, "BritishThermalUnit", "WattHour")
	checkConvertFunction(t, Functions.ConvertForce, "convert_force", 15, "Decanewton", "TonneForce")
	checkConvertFunction(t, Functions.ConvertLength, "convert_length", 36, "Angstrom", "Yard")
	checkConvertFunction(t, Functions.ConvertMass, "convert_mass", 25, "Centigram", "Tonne")
	checkConvertFunction(t, Functions.ConvertSpeed, "convert_speed", 32, "CentimeterPerHour", "YardPerSecond")
	checkConvertFunction(t, Functions.ConvertTemperature, "convert_temperature", 10, "DegreeCelsius", "SolarTemperature")
	checkConvertFunction(t, Functions.ConvertVolume, "convert_volume", 52, "AcreFoot", "UsTeaspoon")
}

func checkSourceReferenceFunctions(t *testing.T) {
	t.Helper()
	checkSourceReferenceFunction(t, Functions.Cluster, "cluster", ReturnTypeParameter0Cluster, 1, 1, []string{"name"}, []*TypeSymbol{ScalarTypes.String.Type()}, []bool{false})
	checkSourceReferenceFunction(t, Functions.Database, "database", ReturnTypeParameter0Database, 0, 1, []string{"name"}, []*TypeSymbol{ScalarTypes.String.Type()}, []bool{true})
	checkSourceReferenceFunction(t, Functions.Table, "table", ReturnTypeParameter0Table, 1, 2, []string{"name", "query_data_scope"}, []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.String.Type()}, []bool{false, true})
	checkSourceReferenceFunction(t, Functions.MaterializedView, "materialized_view", ReturnTypeParameter0MaterializedView, 1, 2, []string{"name", "max_age"}, []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.TimeSpan.Type()}, []bool{false, true})
	checkSourceReferenceFunction(t, Functions.EntityGroup, "entity_group", ReturnTypeParameter0EntityGroup, 1, 1, []string{"name"}, []*TypeSymbol{ScalarTypes.String.Type()}, []bool{false})

	externalTableSignatures := Functions.ExternalTable.Signatures()
	if Functions.ExternalTable.SymbolName() != "external_table" || Functions.ExternalTable.IsConstantFoldable() || Functions.ExternalTable.ResultNameKind() != ResultNameNone || len(externalTableSignatures) != 2 {
		t.Fatalf("external_table metadata = name:%s fold:%t result:%s signatures:%d", Functions.ExternalTable.SymbolName(), Functions.ExternalTable.IsConstantFoldable(), Functions.ExternalTable.ResultNameKind(), len(externalTableSignatures))
	}
	checkSourceReferenceSignature(t, externalTableSignatures[0], "external_table", ReturnTypeParameter0ExternalTable, 1, 2, []string{"name", "mapping"}, []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.String.Type()}, []bool{false, true})
	checkSourceReferenceSignature(t, externalTableSignatures[1], "external_table", ReturnTypeParameter0ExternalTable, 2, 2, []string{"name", "max_age"}, []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.TimeSpan.Type()}, []bool{false, false})

	storedQueryResultSignatures := Functions.StoredQueryResult.Signatures()
	if Functions.StoredQueryResult.SymbolName() != "stored_query_result" || Functions.StoredQueryResult.IsConstantFoldable() || Functions.StoredQueryResult.ResultNameKind() != ResultNameNone || len(storedQueryResultSignatures) != 3 {
		t.Fatalf("stored_query_result metadata = name:%s fold:%t result:%s signatures:%d", Functions.StoredQueryResult.SymbolName(), Functions.StoredQueryResult.IsConstantFoldable(), Functions.StoredQueryResult.ResultNameKind(), len(storedQueryResultSignatures))
	}
	checkSourceReferenceSignature(t, storedQueryResultSignatures[0], "stored_query_result", ReturnTypeParameter0StoredQueryResult, 1, 1, []string{"name"}, []*TypeSymbol{ScalarTypes.String.Type()}, []bool{false})
	checkSourceReferenceSignature(t, storedQueryResultSignatures[1], "stored_query_result", ReturnTypeParameter0StoredQueryResult, 2, 2, []string{"name", "index"}, []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.Long.Type()}, []bool{false, false})
	checkSourceReferenceSignature(t, storedQueryResultSignatures[2], "stored_query_result", ReturnTypeParameter0StoredQueryResult, 2, 2, []string{"name", "table_name"}, []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.String.Type()}, []bool{false, false})

	graphSignatures := Functions.Graph.Signatures()
	if Functions.Graph.SymbolName() != "graph" || Functions.Graph.IsConstantFoldable() || Functions.Graph.ResultNameKind() != ResultNameNone || len(graphSignatures) != 2 {
		t.Fatalf("graph metadata = name:%s fold:%t result:%s signatures:%d", Functions.Graph.SymbolName(), Functions.Graph.IsConstantFoldable(), Functions.Graph.ResultNameKind(), len(graphSignatures))
	}
	checkSourceReferenceSignature(t, graphSignatures[0], "graph", ReturnTypeParameter0Graph, 1, 2, []string{"name", "snapshot"}, []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.String.Type()}, []bool{false, true})
	checkSourceReferenceSignature(t, graphSignatures[1], "graph", ReturnTypeParameter0Graph, 2, 2, []string{"name", "volatile"}, []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.Bool.Type()}, []bool{false, false})
}

func checkSourceReferenceFunction(t *testing.T, fn *FunctionSymbol, name string, returnKind ReturnTypeKind, minArgs int, maxArgs int, parameterNames []string, parameterTypes []*TypeSymbol, optional []bool) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone || fn.IsHidden() {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s hidden:%t", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.IsHidden())
	}
	signatures := fn.Signatures()
	if len(signatures) != 1 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	checkSourceReferenceSignature(t, signatures[0], name, returnKind, minArgs, maxArgs, parameterNames, parameterTypes, optional)
}

func checkSourceReferenceSignature(t *testing.T, signature *Signature, name string, returnKind ReturnTypeKind, minArgs int, maxArgs int, parameterNames []string, parameterTypes []*TypeSymbol, optional []bool) {
	t.Helper()
	if signature.ReturnKind != returnKind || signature.DeclaredReturnType != nil || signature.DeclaredReturnSymbol != nil || signature.MinArgumentCount != minArgs || signature.MaxArgumentCount != maxArgs {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	if len(signature.Parameters) != len(parameterNames) || len(parameterTypes) != len(parameterNames) || len(optional) != len(parameterNames) {
		t.Fatalf("%s test parameter shape mismatch", name)
	}
	for i, parameter := range signature.Parameters {
		if parameter.Name != parameterNames[i] || parameter.TypeKind != ParameterTypeDeclared || len(parameter.DeclaredTypes) != 1 || parameter.DeclaredTypes[0] != parameterTypes[i] || parameter.IsOptional() != optional[i] {
			t.Fatalf("%s parameter %d = %#v", name, i, parameter)
		}
	}
}

func checkStringFunctions(t *testing.T) {
	t.Helper()
	checkSimpleStringKindFunction(t, Functions.Strcat, "strcat", ScalarTypes.String.Type(), ResultNameNone, 1, 64, []string{"arg"}, []ParameterTypeKind{ParameterTypeScalar})
	checkSimpleStringKindFunction(t, Functions.StrcatArray, "strcat_array", ScalarTypes.String.Type(), ResultNameNone, 2, 2, []string{"array", "delimiter"}, []ParameterTypeKind{ParameterTypeDynamicArray, ParameterTypeScalar})
	checkSimpleStringKindFunction(t, Functions.ArrayStrcat, "array_strcat", ScalarTypes.String.Type(), ResultNameNone, 2, 2, []string{"array", "delimiter"}, []ParameterTypeKind{ParameterTypeDynamicArray, ParameterTypeScalar})
	checkSimpleStringKindFunction(t, Functions.StrcatDelim, "strcat_delim", ScalarTypes.String.Type(), ResultNameNone, 3, 65, []string{"delimiter", "arg"}, []ParameterTypeKind{ParameterTypeScalar, ParameterTypeScalar})
	if arg := Functions.StrcatDelim.Signatures()[0].Parameters[1]; arg.MinOccurring != 2 || arg.MaxOccurring != 64 {
		t.Fatalf("strcat_delim arg parameter = %#v", arg)
	}
	checkSimpleStringKindFunction(t, Functions.Strcmp, "strcmp", ScalarTypes.Long.Type(), ResultNameNone, 2, 2, []string{"s1", "s2"}, []ParameterTypeKind{ParameterTypeScalar, ParameterTypeScalar})
	checkStringStrrepFunction(t)
	checkSimpleStringKindFunction(t, Functions.Strlen, "strlen", ScalarTypes.Long.Type(), ResultNameNameAndFirstArgument, 1, 1, []string{"string"}, []ParameterTypeKind{ParameterTypeStringOrDynamic})
	checkSimpleStringKindFunction(t, Functions.StringSize, "string_size", ScalarTypes.Long.Type(), ResultNameNameAndFirstArgument, 1, 1, []string{"string"}, []ParameterTypeKind{ParameterTypeStringOrDynamic})
	checkSimpleStringKindFunction(t, Functions.ToUpper, "toupper", ScalarTypes.String.Type(), ResultNameNone, 1, 1, []string{"value"}, []ParameterTypeKind{ParameterTypeScalar})
	checkSimpleStringKindFunction(t, Functions.ToLower, "tolower", ScalarTypes.String.Type(), ResultNameNone, 1, 1, []string{"value"}, []ParameterTypeKind{ParameterTypeScalar})
	checkObsoleteStringFunction(t, Functions.ToUtf8Deprecated, "to_utf8", "unicode_codepoints_from_string")
	checkSimpleStringKindFunction(t, Functions.UnicodeCodepointsFromString, "unicode_codepoints_from_string", ScalarTypes.DynamicArray.Type(), ResultNameNone, 1, 1, []string{"value"}, []ParameterTypeKind{ParameterTypeScalar})
	checkSimpleStringKindFunction(t, Functions.Substring, "substring", ScalarTypes.String.Type(), ResultNameNone, 2, 3, []string{"string", "start", "length"}, []ParameterTypeKind{ParameterTypeScalar, ParameterTypeInteger, ParameterTypeInteger})
	if length := Functions.Substring.Signatures()[0].Parameters[2]; !length.IsOptional() {
		t.Fatalf("substring length parameter = %#v", length)
	}
	checkSimpleStringKindFunction(t, Functions.RegexQuote, "regex_quote", ScalarTypes.String.Type(), ResultNameNone, 1, 1, []string{"string"}, []ParameterTypeKind{ParameterTypeScalar})
	checkIndexOfFunction(t, Functions.IndexOf, "indexof", false)
	checkIndexOfFunction(t, Functions.IndexOfRegex, "indexof_regex", true)
	checkSimpleStringKindFunction(t, Functions.HasAnyIndex, "has_any_index", ScalarTypes.Long.Type(), ResultNameNone, 2, 2, []string{"source", "values"}, []ParameterTypeKind{ParameterTypeStringOrDynamic, ParameterTypeDynamicArray})
	checkSimpleStringKindFunction(t, Functions.Reverse, "reverse", ScalarTypes.String.Type(), ResultNameNone, 1, 1, []string{"value"}, []ParameterTypeKind{ParameterTypeScalar})
	checkSplitFunction(t)
	checkSimpleStringDeclaredFunction(t, Functions.ParseCommandLine, "parse_command_line", ScalarTypes.DynamicArrayOfString.Type(), ResultNameNone, []string{"command", "parser"}, []ParameterTypeKind{ParameterTypeScalar, ParameterTypeDeclared})
	checkExtractFunction(t)
	checkExtractAllFunction(t, Functions.ExtractAllDeprecated, "extractall", true)
	checkExtractAllFunction(t, Functions.ExtractAll, "extract_all", false)
	checkExtractJsonFunction(t, Functions.ExtractJsonDeprecated, "extractjson")
	checkExtractJsonFunction(t, Functions.ExtractJson, "extract_json")
	checkObsoleteStringFunction(t, Functions.Replace, "replace", "replace_regex' or 'replace_string")
	checkRegexRewriteFunction(t, Functions.ReplaceRegex, "replace_regex", []string{"source", "lookup_regex", "rewrite_pattern"}, []bool{false, true, true})
	checkSimpleStringDeclaredFunction(t, Functions.ReplaceString, "replace_string", ScalarTypes.String.Type(), ResultNameNone, []string{"text", "lookup", "rewrite"}, []ParameterTypeKind{ParameterTypeDeclared, ParameterTypeDeclared, ParameterTypeDeclared})
	checkSimpleStringDeclaredFunction(t, Functions.ReplaceStrings, "replace_strings", ScalarTypes.String.Type(), ResultNameNone, []string{"text", "lookups", "rewrites"}, []ParameterTypeKind{ParameterTypeDeclared, ParameterTypeDynamicArray, ParameterTypeDynamicArray})
	checkRegexRewriteFunction(t, Functions.TrimStart, "trim_start", []string{"regex", "source"}, []bool{true, false})
	checkRegexRewriteFunction(t, Functions.TrimEnd, "trim_end", []string{"regex", "source"}, []bool{true, false})
	checkRegexRewriteFunction(t, Functions.Trim, "trim", []string{"regex", "source"}, []bool{true, false})
	checkCountOfFunction(t)
	checkSimpleStringDeclaredFunction(t, Functions.Translate, "translate", ScalarTypes.String.Type(), ResultNameNone, []string{"searchList", "replacementList", "source"}, []ParameterTypeKind{ParameterTypeDeclared, ParameterTypeDeclared, ParameterTypeDeclared})
	checkObsoleteStringFunction(t, Functions.MakeStringDeprecated, "make_string", "unicode_codepoints_to_string")
	checkMakeStringFunction(t, Functions.UnicodeCodepointsToString, "unicode_codepoints_to_string")
}

func checkSimpleStringKindFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, resultName ResultNameKind, minArgs int, maxArgs int, parameterNames []string, parameterKinds []ParameterTypeKind) {
	t.Helper()
	checkSimpleStringDeclaredFunction(t, fn, name, returnType, resultName, parameterNames, parameterKinds)
	signature := fn.Signatures()[0]
	if signature.MinArgumentCount != minArgs || signature.MaxArgumentCount != maxArgs {
		t.Fatalf("%s argument count = %d/%d", name, signature.MinArgumentCount, signature.MaxArgumentCount)
	}
}

func checkSimpleStringDeclaredFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, resultName ResultNameKind, parameterNames []string, parameterKinds []ParameterTypeKind) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.IsHidden() || fn.ResultNameKind() != resultName {
		t.Fatalf("%s metadata = name:%s fold:%t hidden:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.IsHidden(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 1 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	signature := signatures[0]
	if signature.DeclaredReturnType != returnType || len(signature.Parameters) != len(parameterNames) {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	for i, parameter := range signature.Parameters {
		if parameter.Name != parameterNames[i] || parameter.TypeKind != parameterKinds[i] {
			t.Fatalf("%s parameter %d = %#v", name, i, parameter)
		}
	}
}

func checkStringStrrepFunction(t *testing.T) {
	t.Helper()
	checkSimpleStringKindFunction(t, Functions.Strrep, "strrep", ScalarTypes.String.Type(), ResultNameNone, 2, 3, []string{"value", "multiplier", "delimiter"}, []ParameterTypeKind{ParameterTypeScalar, ParameterTypeDeclared, ParameterTypeScalar})
	signature := Functions.Strrep.Signatures()[0]
	if signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.Long.Type() || !signature.Parameters[2].IsOptional() {
		t.Fatalf("strrep signature = %#v", signature)
	}
}

func checkObsoleteStringFunction(t *testing.T, fn *FunctionSymbol, name string, alternative string) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsHidden() || fn.Alternative() != alternative || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s obsolete string metadata = name:%s hidden:%t alt:%q fold:%t result:%s", name, fn.SymbolName(), fn.IsHidden(), fn.Alternative(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
}

func checkIndexOfFunction(t *testing.T, fn *FunctionSymbol, name string, regexMatch bool) {
	t.Helper()
	checkSimpleStringKindFunction(t, fn, name, ScalarTypes.Long.Type(), ResultNameNone, 2, 5, []string{"string", "match", "start", "length", "occurrence"}, []ParameterTypeKind{ParameterTypeScalar, ParameterTypeScalar, ParameterTypeInteger, ParameterTypeInteger, ParameterTypeInteger})
	signature := fn.Signatures()[0]
	if regexMatch && signature.Parameters[1].ArgumentKind != ArgumentKindConstant {
		t.Fatalf("%s match parameter = %#v", name, signature.Parameters[1])
	}
	if !signature.Parameters[2].IsOptional() || !signature.Parameters[3].IsOptional() || !signature.Parameters[4].IsOptional() || signature.Parameters[4].ArgumentKind != ArgumentKindConstant {
		t.Fatalf("%s optional parameters = %#v", name, signature.Parameters)
	}
}

func checkSplitFunction(t *testing.T) {
	t.Helper()
	checkSimpleStringKindFunction(t, Functions.Split, "split", ScalarTypes.DynamicArrayOfString.Type(), ResultNameFirstArgument, 2, 3, []string{"source", "delimiter", "requestedIndex"}, []ParameterTypeKind{ParameterTypeScalar, ParameterTypeDeclared, ParameterTypeInteger})
	signature := Functions.Split.Signatures()[0]
	if signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.String.Type() || !signature.Parameters[2].IsOptional() {
		t.Fatalf("split signature = %#v", signature)
	}
}

func checkExtractFunction(t *testing.T) {
	t.Helper()
	fn := Functions.Extract
	if fn.SymbolName() != "extract" || !fn.IsConstantFoldable() || fn.IsHidden() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("extract metadata = name:%s fold:%t hidden:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.IsHidden(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 2 {
		t.Fatalf("extract signatures = %d", len(signatures))
	}
	if signatures[0].DeclaredReturnType != ScalarTypes.String.Type() || signatures[0].Parameters[0].ArgumentKind != ArgumentKindConstant || signatures[0].Parameters[1].DeclaredTypes[0] != ScalarTypes.Long.Type() {
		t.Fatalf("extract first signature = %#v", signatures[0])
	}
	if signatures[1].ReturnKind != ReturnTypeParameterNLiteral || signatures[1].Parameters[3].DeclaredTypes[0] != ScalarTypes.TypeLit.Type() || signatures[1].Parameters[3].ArgumentKind != ArgumentKindLiteral {
		t.Fatalf("extract typed signature = %#v", signatures[1])
	}
}

func checkExtractAllFunction(t *testing.T, fn *FunctionSymbol, name string, obsolete bool) {
	t.Helper()
	if obsolete {
		checkObsoleteStringFunction(t, fn, name, "extract_all")
	} else if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.IsHidden() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t hidden:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.IsHidden(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 2 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	if signatures[0].DeclaredReturnType != ScalarTypes.DynamicArray.Type() || signatures[0].Parameters[0].ArgumentKind != ArgumentKindConstant || signatures[0].Parameters[1].Name != "source" {
		t.Fatalf("%s first signature = %#v", name, signatures[0])
	}
	if signatures[1].Parameters[1].Name != "captureGroups" || signatures[1].Parameters[1].TypeKind != ParameterTypeDynamicArray || signatures[1].Parameters[2].Name != "source" {
		t.Fatalf("%s second signature = %#v", name, signatures[1])
	}
}

func checkExtractJsonFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.IsHidden() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t hidden:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.IsHidden(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 2 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	if signatures[0].DeclaredReturnType != ScalarTypes.String.Type() || signatures[0].Parameters[0].ArgumentKind != ArgumentKindConstant || signatures[0].Parameters[1].Name != "jsonText" {
		t.Fatalf("%s first signature = %#v", name, signatures[0])
	}
	if signatures[1].ReturnKind != ReturnTypeParameterNLiteral || signatures[1].Parameters[2].DeclaredTypes[0] != ScalarTypes.TypeLit.Type() || signatures[1].Parameters[2].ArgumentKind != ArgumentKindLiteral {
		t.Fatalf("%s typed signature = %#v", name, signatures[1])
	}
}

func checkRegexRewriteFunction(t *testing.T, fn *FunctionSymbol, name string, parameterNames []string, constant []bool) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.IsHidden() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t hidden:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.IsHidden(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.String.Type() || len(signature.Parameters) != len(parameterNames) {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	for i, parameter := range signature.Parameters {
		if parameter.Name != parameterNames[i] || parameter.DeclaredTypes[0] != ScalarTypes.String.Type() || (parameter.ArgumentKind == ArgumentKindConstant) != constant[i] {
			t.Fatalf("%s parameter %d = %#v", name, i, parameter)
		}
	}
}

func checkCountOfFunction(t *testing.T) {
	t.Helper()
	checkSimpleStringDeclaredFunction(t, Functions.CountOf, "countof", ScalarTypes.Long.Type(), ResultNameNone, []string{"source", "search", "kind"}, []ParameterTypeKind{ParameterTypeDeclared, ParameterTypeDeclared, ParameterTypeDeclared})
	kind := Functions.CountOf.Signatures()[0].Parameters[2]
	if kind.ArgumentKind != ArgumentKindLiteral || !kind.IsOptional() || !kind.IsCaseSensitive || len(kind.Values) != 2 || kind.Values[0] != "normal" || kind.Values[1] != "regex" {
		t.Fatalf("countof kind parameter = %#v", kind)
	}
}

func checkMakeStringFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	checkSimpleStringKindFunction(t, fn, name, ScalarTypes.String.Type(), ResultNameNone, 1, MaxRepeat, []string{"value"}, []ParameterTypeKind{ParameterTypeIntegerOrArray})
	signature := fn.Signatures()[0]
	if signature.Parameters[0].MinOccurring != 1 || signature.Parameters[0].MaxOccurring != MaxRepeat {
		t.Fatalf("%s value parameter = %#v", name, signature.Parameters[0])
	}
}

func checkTypeConversionFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, valueTypeKind ParameterTypeKind, resultName ResultNameKind) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != resultName {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 1 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	signature := signatures[0]
	if signature.DeclaredReturnType != returnType {
		t.Fatalf("%s return type = %s", name, signature.DeclaredReturnType.SymbolName())
	}
	if len(signature.Parameters) == 0 || signature.Parameters[0].Name != "value" || signature.Parameters[0].TypeKind != valueTypeKind {
		t.Fatalf("%s value parameter = %#v", name, signature.Parameters)
	}
}

func checkParseFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, parameterName string, parameterTypeKind ParameterTypeKind, resultName ResultNameKind) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != resultName {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 1 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	signature := signatures[0]
	if signature.DeclaredReturnType != returnType || len(signature.Parameters) == 0 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	parameter := signature.Parameters[0]
	if parameter.Name != parameterName || parameter.TypeKind != parameterTypeKind {
		t.Fatalf("%s parameter = %#v", name, parameter)
	}
}

func assertDynamicBagProperties(t *testing.T, symbol Symbol, names []string) {
	t.Helper()
	bag, ok := symbol.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("return symbol = %T, want DynamicBagSymbol", symbol)
	}
	properties := bag.Properties()
	if len(properties) != len(names) {
		t.Fatalf("bag properties = %#v", properties)
	}
	for i, name := range names {
		if properties[i].SymbolName() != name {
			t.Fatalf("bag property %d = %q, want %q", i, properties[i].SymbolName(), name)
		}
	}
}

func checkEncodingFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, parameterName string, parameterTypeKind ParameterTypeKind) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 1 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	signature := signatures[0]
	if signature.DeclaredReturnType != returnType || signature.MinArgumentCount != 1 || signature.MaxArgumentCount != 1 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	parameter := signature.Parameters[0]
	if parameter.Name != parameterName || parameter.TypeKind != parameterTypeKind {
		t.Fatalf("%s parameter = %#v", name, parameter)
	}
}

func checkDateTimeFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, parameterName string, resultName ResultNameKind) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != resultName {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 1 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	signature := signatures[0]
	if signature.DeclaredReturnType != returnType {
		t.Fatalf("%s return type = %s", name, signature.DeclaredReturnType.SymbolName())
	}
	if parameterName == "" {
		if len(signature.Parameters) != 0 || signature.MinArgumentCount != 0 || signature.MaxArgumentCount != 0 {
			t.Fatalf("%s signature = %#v", name, signature)
		}
		return
	}
	if len(signature.Parameters) == 0 || signature.Parameters[0].Name != parameterName {
		t.Fatalf("%s parameter = %#v", name, signature.Parameters)
	}
}

func checkDateTimePartFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, parts []string, maxArgs int) {
	t.Helper()
	checkDateTimeFunction(t, fn, name, returnType, "part", ResultNameNone)
	signature := fn.Signatures()[0]
	if signature.MinArgumentCount != maxArgs || signature.MaxArgumentCount != maxArgs {
		t.Fatalf("%s argument count = %d/%d", name, signature.MinArgumentCount, signature.MaxArgumentCount)
	}
	parameter := signature.Parameters[0]
	if parameter.ArgumentKind != ArgumentKindLiteral || len(parameter.Values) != len(parts) {
		t.Fatalf("%s part parameter = %#v", name, parameter)
	}
	if parameter.Values[0] != parts[0] || parameter.Values[len(parameter.Values)-1] != parts[len(parts)-1] {
		t.Fatalf("%s part value bounds = %#v", name, parameter.Values)
	}
}

func checkMakeDatetime(t *testing.T) {
	t.Helper()
	fn := Functions.MakeDatetime
	if fn.SymbolName() != "make_datetime" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameOnlyArgument {
		t.Fatalf("make_datetime metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 2 {
		t.Fatalf("make_datetime signatures = %d", len(signatures))
	}
	if signatures[0].DeclaredReturnType != ScalarTypes.DateTime.Type() || signatures[0].MinArgumentCount != 3 || signatures[0].MaxArgumentCount != 6 {
		t.Fatalf("make_datetime first signature = %#v", signatures[0])
	}
	if signatures[0].Parameters[3].Name != "hour" || !signatures[0].Parameters[3].IsOptional() {
		t.Fatalf("make_datetime hour parameter = %#v", signatures[0].Parameters[3])
	}
	if signatures[1].Parameters[0].Name != "value" || signatures[1].Parameters[0].TypeKind != ParameterTypeScalar {
		t.Fatalf("make_datetime value signature = %#v", signatures[1])
	}
}

func checkMakeTimespan(t *testing.T) {
	t.Helper()
	fn := Functions.MakeTimespan
	if fn.SymbolName() != "make_timespan" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameOnlyArgument {
		t.Fatalf("make_timespan metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 3 {
		t.Fatalf("make_timespan signatures = %d", len(signatures))
	}
	if signatures[0].DeclaredReturnType != ScalarTypes.TimeSpan.Type() || signatures[0].MinArgumentCount != 4 || signatures[0].MaxArgumentCount != 4 {
		t.Fatalf("make_timespan first signature = %#v", signatures[0])
	}
	if signatures[1].MinArgumentCount != 2 || signatures[1].MaxArgumentCount != 3 || !signatures[1].Parameters[2].IsOptional() {
		t.Fatalf("make_timespan second signature = %#v", signatures[1])
	}
	if signatures[2].Parameters[0].Name != "value" || signatures[2].Parameters[0].TypeKind != ParameterTypeScalar {
		t.Fatalf("make_timespan value signature = %#v", signatures[2])
	}
}

func checkOptionalOffset(t *testing.T, fn *FunctionSymbol, offsetType *TypeSymbol) {
	t.Helper()
	signature := fn.Signatures()[0]
	if len(signature.Parameters) != 2 {
		t.Fatalf("%s parameters = %#v", fn.SymbolName(), signature.Parameters)
	}
	parameter := signature.Parameters[1]
	if parameter.Name != "offset" || !parameter.IsOptional() || parameter.DeclaredTypes[0] != offsetType {
		t.Fatalf("%s offset parameter = %#v", fn.SymbolName(), parameter)
	}
}

func checkHashFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, hidden bool, minArgs int, maxArgs int) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone || fn.IsHidden() != hidden {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s hidden:%t", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.IsHidden())
	}
	signatures := fn.Signatures()
	if len(signatures) != 1 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	signature := signatures[0]
	if signature.DeclaredReturnType != returnType || signature.MinArgumentCount != minArgs || signature.MaxArgumentCount != maxArgs {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	if len(signature.Parameters) == 0 {
		t.Fatalf("%s parameters = %#v", name, signature.Parameters)
	}
	first := signature.Parameters[0]
	if name == "__hash_many_crc32" {
		if first.Name != "arg" || first.TypeKind != ParameterTypeScalar || !first.IsRepeatable() {
			t.Fatalf("%s first parameter = %#v", name, first)
		}
		return
	}
	if first.Name != "source" || first.TypeKind != ParameterTypeNotDynamic && name != "hash_combine" {
		t.Fatalf("%s first parameter = %#v", name, first)
	}
	if name == "hash_combine" && (first.TypeKind != ParameterTypeScalar || first.MinOccurring != 2 || first.MaxOccurring != MaxRepeat) {
		t.Fatalf("%s source parameter = %#v", name, first)
	}
	if maxArgs == 2 && len(signature.Parameters) == 2 {
		mod := signature.Parameters[1]
		if mod.Name != "mod" || mod.TypeKind != ParameterTypeInteger || !mod.IsOptional() {
			t.Fatalf("%s mod parameter = %#v", name, mod)
		}
	}
}

func checkConditionalFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 1 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	signature := signatures[0]
	if signature.ReturnKind != ReturnTypeCommonNonDynamic || signature.DeclaredReturnType != nil || signature.MinArgumentCount != 3 || signature.MaxArgumentCount != 3 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	if signature.Parameters[0].Name != "if" || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.Bool.Type() {
		t.Fatalf("%s if parameter = %#v", name, signature.Parameters[0])
	}
	if signature.Parameters[1].Name != "then" || signature.Parameters[1].TypeKind != ParameterTypeCommonScalarOrDynamic ||
		signature.Parameters[2].Name != "else" || signature.Parameters[2].TypeKind != ParameterTypeCommonScalarOrDynamic {
		t.Fatalf("%s branch parameters = %#v", name, signature.Parameters)
	}
}

func checkCaseFunction(t *testing.T) {
	t.Helper()
	fn := Functions.Case
	if fn.SymbolName() != "case" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("case metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 1 {
		t.Fatalf("case signatures = %d", len(signatures))
	}
	signature := signatures[0]
	if signature.ReturnKind != ReturnTypeCommonNonDynamic || signature.Layout != ParameterLayoutBlockRepeating || signature.MinArgumentCount != 3 {
		t.Fatalf("case signature = %#v", signature)
	}
	if !signature.IsValidArgumentCount(3) || signature.IsValidArgumentCount(4) || !signature.IsValidArgumentCount(5) {
		t.Fatalf("case block-repeating argument validation failed")
	}
	if signature.Parameters[0].Name != "predicate" || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.Bool.Type() || !signature.Parameters[0].IsRepeatable() {
		t.Fatalf("case predicate parameter = %#v", signature.Parameters[0])
	}
	if signature.Parameters[1].Name != "then" || signature.Parameters[1].TypeKind != ParameterTypeCommonScalarOrDynamic || !signature.Parameters[1].IsRepeatable() {
		t.Fatalf("case then parameter = %#v", signature.Parameters[1])
	}
	if signature.Parameters[2].Name != "else" || signature.Parameters[2].TypeKind != ParameterTypeCommonScalarOrDynamic {
		t.Fatalf("case else parameter = %#v", signature.Parameters[2])
	}
}

func checkAssertFunction(t *testing.T) {
	t.Helper()
	fn := Functions.Assert
	if fn.SymbolName() != "assert" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("assert metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Bool.Type() || signature.MinArgumentCount != 2 || signature.MaxArgumentCount != 2 {
		t.Fatalf("assert signature = %#v", signature)
	}
	if signature.Parameters[0].Name != "predicate" || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.Bool.Type() ||
		signature.Parameters[1].Name != "message" || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.String.Type() {
		t.Fatalf("assert parameters = %#v", signature.Parameters)
	}
}

func checkBinFloorFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 4 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	if signatures[0].ReturnKind != ReturnTypeWidest || signatures[0].Parameters[0].TypeKind != ParameterTypeNumber || signatures[0].Parameters[1].TypeKind != ParameterTypeNumber {
		t.Fatalf("%s numeric signature = %#v", name, signatures[0])
	}
	if signatures[1].DeclaredReturnType != ScalarTypes.TimeSpan.Type() ||
		signatures[1].Parameters[0].DeclaredTypes[0] != ScalarTypes.TimeSpan.Type() ||
		signatures[1].Parameters[1].DeclaredTypes[0] != ScalarTypes.TimeSpan.Type() {
		t.Fatalf("%s timespan signature = %#v", name, signatures[1])
	}
	if signatures[2].DeclaredReturnType != ScalarTypes.DateTime.Type() ||
		signatures[2].Parameters[0].DeclaredTypes[0] != ScalarTypes.DateTime.Type() ||
		signatures[2].Parameters[1].DeclaredTypes[0] != ScalarTypes.DateTime.Type() {
		t.Fatalf("%s datetime signature = %#v", name, signatures[2])
	}
	if signatures[3].DeclaredReturnType != ScalarTypes.DateTime.Type() ||
		signatures[3].Parameters[0].DeclaredTypes[0] != ScalarTypes.DateTime.Type() ||
		signatures[3].Parameters[1].DeclaredTypes[0] != ScalarTypes.TimeSpan.Type() {
		t.Fatalf("%s datetime-timespan signature = %#v", name, signatures[3])
	}
}

func checkBinAtFunction(t *testing.T) {
	t.Helper()
	fn := Functions.BinAt
	if fn.SymbolName() != "bin_at" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument {
		t.Fatalf("bin_at metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 4 {
		t.Fatalf("bin_at signatures = %d", len(signatures))
	}
	if signatures[0].ReturnKind != ReturnTypeWidest || signatures[0].MinArgumentCount != 3 || signatures[0].MaxArgumentCount != 3 {
		t.Fatalf("bin_at numeric signature = %#v", signatures[0])
	}
	if signatures[1].DeclaredReturnType != ScalarTypes.TimeSpan.Type() ||
		signatures[1].Parameters[1].TypeKind != ParameterTypeNumber ||
		signatures[1].Parameters[2].DeclaredTypes[0] != ScalarTypes.TimeSpan.Type() {
		t.Fatalf("bin_at timespan-number signature = %#v", signatures[1])
	}
	if signatures[2].DeclaredReturnType != ScalarTypes.TimeSpan.Type() ||
		signatures[2].Parameters[1].DeclaredTypes[0] != ScalarTypes.TimeSpan.Type() {
		t.Fatalf("bin_at timespan signature = %#v", signatures[2])
	}
	if signatures[3].DeclaredReturnType != ScalarTypes.DateTime.Type() ||
		signatures[3].Parameters[1].TypeKind != ParameterTypeSummable ||
		signatures[3].Parameters[2].DeclaredTypes[0] != ScalarTypes.DateTime.Type() {
		t.Fatalf("bin_at datetime signature = %#v", signatures[3])
	}
}

func checkBinAutoFunction(t *testing.T) {
	t.Helper()
	fn := Functions.BinAuto
	if fn.SymbolName() != "bin_auto" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument || !fn.IsHidden() {
		t.Fatalf("bin_auto metadata = name:%s fold:%t result:%s hidden:%t", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.IsHidden())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.CustomReturnType == nil || signature.Tabularity != TabularityScalar {
		t.Fatalf("bin_auto signature = %#v", signature)
	}
	if signature.Parameters[0].Name != "value" || signature.Parameters[0].TypeKind != ParameterTypeSummable {
		t.Fatalf("bin_auto value parameter = %#v", signature.Parameters[0])
	}
	if got := signature.CustomReturnType(CustomReturnTypeContext{ArgumentTypes: []*TypeSymbol{ScalarTypes.Int.Type()}}); got != ScalarTypes.Long.Type() {
		t.Fatalf("bin_auto int return = %#v", got)
	}
	if got := signature.CustomReturnType(CustomReturnTypeContext{ArgumentTypes: []*TypeSymbol{ScalarTypes.Real.Type()}}); got != ScalarTypes.Real.Type() {
		t.Fatalf("bin_auto real return = %#v", got)
	}
	if got := signature.CustomReturnType(CustomReturnTypeContext{ArgumentTypes: []*TypeSymbol{ScalarTypes.DateTime.Type()}}); got != ScalarTypes.DateTime.Type() {
		t.Fatalf("bin_auto datetime return = %#v", got)
	}
	withNamedOptions := CustomReturnTypeContext{
		ArgumentTypes: []*TypeSymbol{ScalarTypes.Int.Type()},
		ResultTypeByName: func(name string) *TypeSymbol {
			switch name {
			case "query_bin_auto_size", "query_bin_auto_at":
				return ScalarTypes.Real.Type()
			default:
				return nil
			}
		},
	}
	if got := signature.CustomReturnType(withNamedOptions); got != ScalarTypes.Real.Type() {
		t.Fatalf("bin_auto named option return = %#v", got)
	}
}

func checkBoolStateFunctions(t *testing.T) {
	t.Helper()
	checkBoolFunction(t, Functions.Not, "not", "expression", ParameterTypeDeclared, false)
	checkBoolFunction(t, Functions.NotNullDeprecated, "notnull", "expression", ParameterTypeScalar, true)
	checkBoolFunction(t, Functions.IsNotNull, "isnotnull", "expression", ParameterTypeScalar, false)
	checkBoolFunction(t, Functions.IsNull, "isnull", "expression", ParameterTypeScalar, false)
	checkBoolFunction(t, Functions.NotEmptyDeprecated, "notempty", "value", ParameterTypeScalar, true)
	checkBoolFunction(t, Functions.IsAscii, "isascii", "value", ParameterTypeScalar, false)
	checkBoolFunction(t, Functions.IsUtf8, "isutf8", "value", ParameterTypeScalar, false)
	checkBoolFunction(t, Functions.IsNotEmpty, "isnotempty", "value", ParameterTypeScalar, false)
	checkBoolFunction(t, Functions.IsEmpty, "isempty", "value", ParameterTypeScalar, false)
	if Functions.NotNullDeprecated.Alternative() != "isnotnull" || Functions.NotEmptyDeprecated.Alternative() != "isnotempty" {
		t.Fatalf("bool obsolete alternatives = %q %q", Functions.NotNullDeprecated.Alternative(), Functions.NotEmptyDeprecated.Alternative())
	}
	if Functions.IsColumnExists.SymbolName() != "iscolumnexists" || !Functions.IsColumnExists.IsHidden() || Functions.IsColumnExists.IsConstantFoldable() {
		t.Fatalf("iscolumnexists metadata = name:%s hidden:%t fold:%t", Functions.IsColumnExists.SymbolName(), Functions.IsColumnExists.IsHidden(), Functions.IsColumnExists.IsConstantFoldable())
	}
	columnExistsSignature := Functions.IsColumnExists.Signatures()[0]
	if columnExistsSignature.DeclaredReturnType != ScalarTypes.Bool.Type() ||
		columnExistsSignature.Parameters[0].Name != "column_name" ||
		columnExistsSignature.Parameters[0].ArgumentKind != ArgumentKindConstant {
		t.Fatalf("iscolumnexists signature = %#v", columnExistsSignature)
	}
	checkColumnIfExistsFunction(t, Functions.ColumnIfExistsDeprecated, "columnifexists", true)
	checkColumnIfExistsFunction(t, Functions.ColumnIfExists, "column_ifexists", false)
	checkBoolFunction(t, Functions.Around, "around", "value", ParameterTypeScalar, false)
	aroundSignature := Functions.Around.Signatures()[0]
	if aroundSignature.MinArgumentCount != 3 || aroundSignature.MaxArgumentCount != 3 {
		t.Fatalf("around argument count = %#v", aroundSignature)
	}
}

func checkBoolFunction(t *testing.T, fn *FunctionSymbol, name string, firstParameter string, firstTypeKind ParameterTypeKind, hidden bool) {
	t.Helper()
	if fn.SymbolName() != name || fn.ResultNameKind() != ResultNameNone || fn.IsConstantFoldable() || fn.IsHidden() != hidden {
		t.Fatalf("%s metadata = name:%s result:%s fold:%t hidden:%t", name, fn.SymbolName(), fn.ResultNameKind(), fn.IsConstantFoldable(), fn.IsHidden())
	}
	signatures := fn.Signatures()
	if len(signatures) != 1 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	signature := signatures[0]
	if signature.DeclaredReturnType != ScalarTypes.Bool.Type() {
		t.Fatalf("%s return type = %#v", name, signature.DeclaredReturnType)
	}
	if signature.Parameters[0].Name != firstParameter || signature.Parameters[0].TypeKind != firstTypeKind {
		t.Fatalf("%s first parameter = %#v", name, signature.Parameters[0])
	}
}

func checkColumnIfExistsFunction(t *testing.T, fn *FunctionSymbol, name string, hidden bool) {
	t.Helper()
	if fn.SymbolName() != name || fn.ResultNameKind() != ResultNameFirstArgumentValueIfColumn || fn.IsConstantFoldable() || fn.IsHidden() != hidden {
		t.Fatalf("%s metadata = name:%s result:%s fold:%t hidden:%t", name, fn.SymbolName(), fn.ResultNameKind(), fn.IsConstantFoldable(), fn.IsHidden())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeParameter1 || signature.MinArgumentCount != 2 || signature.MaxArgumentCount != 2 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	if signature.Parameters[0].Name != "column_name" || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.String.Type() || signature.Parameters[0].ArgumentKind != ArgumentKindConstant {
		t.Fatalf("%s column_name parameter = %#v", name, signature.Parameters[0])
	}
	if signature.Parameters[1].Name != "defaultValue" || signature.Parameters[1].TypeKind != ParameterTypeScalar {
		t.Fatalf("%s defaultValue parameter = %#v", name, signature.Parameters[1])
	}
	if hidden && fn.Alternative() != "column_ifexists" {
		t.Fatalf("%s alternative = %q", name, fn.Alternative())
	}
}

func checkBitwiseFunctions(t *testing.T) {
	t.Helper()
	checkBitwiseFunction(t, Functions.BinaryAnd, "binary_and", []string{"value1", "value2"})
	checkBitwiseFunction(t, Functions.BinaryOr, "binary_or", []string{"value1", "value2"})
	checkBitwiseFunction(t, Functions.BinaryXor, "binary_xor", []string{"value1", "value2"})
	checkBitwiseFunction(t, Functions.BinaryNot, "binary_not", []string{"value"})
	checkBitwiseFunction(t, Functions.BinaryShiftRight, "binary_shift_right", []string{"value", "shift"})
	checkBitwiseFunction(t, Functions.BinaryShiftLeft, "binary_shift_left", []string{"value", "shift"})
	checkBitwiseFunction(t, Functions.BitsetCountOnes, "bitset_count_ones", []string{"value"})
}

func checkBitwiseFunction(t *testing.T, fn *FunctionSymbol, name string, parameterNames []string) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument || fn.IsHidden() {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s hidden:%t", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.IsHidden())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Long.Type() || signature.MinArgumentCount != len(parameterNames) || signature.MaxArgumentCount != len(parameterNames) {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	for i, parameterName := range parameterNames {
		parameter := signature.Parameters[i]
		if parameter.Name != parameterName || parameter.TypeKind != ParameterTypeInteger {
			t.Fatalf("%s parameter %d = %#v", name, i, parameter)
		}
	}
}

func checkDynamicArrayFunctions(t *testing.T) {
	t.Helper()
	checkTreePathFunction(t)
	checkRepeatFunction(t)
	checkArrayLengthFunction(t, Functions.ArrayLengthDeprecated, "arraylength", true)
	checkArrayLengthFunction(t, Functions.ArrayLength, "array_length", false)
	checkRangeFunction(t)
	checkArrayConcatFunction(t)
	checkArrayIifFunction(t, Functions.ArrayIif, "array_iif", true)
	checkArrayIifFunction(t, Functions.ArrayIff, "array_iff", false)
	checkArrayIndexOfFunction(t)
	checkReturnParameter0Function(t, Functions.ArraySlice, "array_slice", []string{"array", "start", "end"}, []ParameterTypeKind{ParameterTypeDynamicArray, ParameterTypeInteger, ParameterTypeInteger}, false)
	checkArraySplitFunction(t)
	checkReturnParameter0Function(t, Functions.ArrayShiftLeft, "array_shift_left", []string{"array", "shift_count", "default_value"}, []ParameterTypeKind{ParameterTypeDynamicArray, ParameterTypeInteger, ParameterTypeScalar}, true)
	checkReturnParameter0Function(t, Functions.ArrayShiftRight, "array_shift_right", []string{"array", "shift_count", "default_value"}, []ParameterTypeKind{ParameterTypeDynamicArray, ParameterTypeInteger, ParameterTypeScalar}, true)
	checkReturnParameter0Function(t, Functions.ArrayReverse, "array_reverse", []string{"value"}, []ParameterTypeKind{ParameterTypeDynamicArray}, false)
	checkReturnParameter0Function(t, Functions.ArrayRotateLeft, "array_rotate_left", []string{"array", "rotate_count"}, []ParameterTypeKind{ParameterTypeDynamicArray, ParameterTypeInteger}, false)
	checkReturnParameter0Function(t, Functions.ArrayRotateRight, "array_rotate_right", []string{"array", "rotate_count"}, []ParameterTypeKind{ParameterTypeDynamicArray, ParameterTypeInteger}, false)
	checkArraySortFunction(t, Functions.ArraySortAsc, "array_sort_asc")
	checkArraySortFunction(t, Functions.ArraySortDesc, "array_sort_desc")
	checkDynamicArrayDeclaredFunction(t, Functions.BagKeys, "bag_keys", ScalarTypes.DynamicArrayOfString.Type(), []string{"object"}, []ParameterTypeKind{ParameterTypeDynamicBag})
	checkDynamicArrayDeclaredFunction(t, Functions.Zip, "zip", ScalarTypes.DynamicArrayOfArray.Type(), []string{"array"}, []ParameterTypeKind{ParameterTypeDynamicArray})
	if parameter := Functions.Zip.Signatures()[0].Parameters[0]; parameter.MinOccurring != 2 || parameter.MaxOccurring != 64 {
		t.Fatalf("zip array occurrence = %#v", parameter)
	}
	checkBagPackFunction(t, Functions.Pack, "pack", true)
	checkBagPackFunction(t, Functions.PackDictionary, "pack_dictionary", true)
	checkBagPackFunction(t, Functions.BagPack, "bag_pack", false)
	checkBagPackColumnsFunction(t)
	checkPackAllFunction(t)
	checkPackArrayFunction(t)
	checkSetFunctions(t)
	checkDynamicBagContinuationFunctions(t)
}

func checkTreePathFunction(t *testing.T) {
	t.Helper()
	fn := Functions.TreePath
	if fn.SymbolName() != "treepath" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNamePrefixAndFirstArgument || fn.ResultNamePrefix() != "tree" {
		t.Fatalf("treepath metadata = name:%s fold:%t result:%s prefix:%q", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.ResultNamePrefix())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnSymbol != ScalarTypes.DynamicArrayOfString || signature.DeclaredReturnType != ScalarTypes.DynamicArrayOfString.Type() {
		t.Fatalf("treepath return = %#v / %#v", signature.DeclaredReturnSymbol, signature.DeclaredReturnType)
	}
	if signature.Parameters[0].Name != "object" || signature.Parameters[0].TypeKind != ParameterTypeDynamicBag {
		t.Fatalf("treepath parameter = %#v", signature.Parameters[0])
	}
}

func checkRepeatFunction(t *testing.T) {
	t.Helper()
	fn := Functions.Repeat
	if fn.SymbolName() != "repeat" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNamePrefixAndFirstArgument || fn.ResultNamePrefix() != "repeat" {
		t.Fatalf("repeat metadata = name:%s fold:%t result:%s prefix:%q", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.ResultNamePrefix())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeParameter0Array || signature.Parameters[0].TypeKind != ParameterTypeScalar || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.Long.Type() {
		t.Fatalf("repeat signature = %#v", signature)
	}
}

func checkArrayLengthFunction(t *testing.T, fn *FunctionSymbol, name string, hidden bool) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone || fn.IsHidden() != hidden {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s hidden:%t", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.IsHidden())
	}
	if hidden && fn.Alternative() != "array_length" {
		t.Fatalf("%s alternative = %q", name, fn.Alternative())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Long.Type() || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkRangeFunction(t *testing.T) {
	t.Helper()
	signature := Functions.Range.Signatures()[0]
	if Functions.Range.SymbolName() != "range" || !Functions.Range.IsConstantFoldable() || Functions.Range.ResultNameKind() != ResultNameNone || signature.ReturnKind != ReturnTypeParameter0Array {
		t.Fatalf("range metadata/signature = %#v", signature)
	}
	if signature.MinArgumentCount != 2 || signature.MaxArgumentCount != 3 || len(signature.Parameters) != 3 || !signature.Parameters[2].IsOptional() {
		t.Fatalf("range argument count = %#v", signature)
	}
	for i, name := range []string{"start", "stop", "step"} {
		if signature.Parameters[i].Name != name || signature.Parameters[i].TypeKind != ParameterTypeSummable {
			t.Fatalf("range parameter %d = %#v", i, signature.Parameters[i])
		}
	}
}

func checkArrayConcatFunction(t *testing.T) {
	t.Helper()
	fn := Functions.ArrayConcat
	if fn.SymbolName() != "array_concat" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("array_concat metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.CustomReturnType == nil || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.Parameters[0].MaxOccurring != 64 {
		t.Fatalf("array_concat signature = %#v", signature)
	}
	if got := signature.CustomReturnType(CustomReturnTypeContext{}); got != ScalarTypes.DynamicArray.Type() {
		t.Fatalf("array_concat fallback return = %#v", got)
	}
	if got := signature.CustomReturnType(CustomReturnTypeContext{
		Signature:          signature,
		ArgumentParameters: []*Parameter{signature.Parameters[0], signature.Parameters[0]},
		ArgumentTypes:      []*TypeSymbol{ScalarTypes.DynamicArrayOfString.Type(), ScalarTypes.DynamicArrayOfString.Type()},
	}); got != ScalarTypes.DynamicArrayOfString.Type() {
		t.Fatalf("array_concat common string-array return = %#v", got)
	}
	if got := signature.CustomReturnType(CustomReturnTypeContext{
		Signature:          signature,
		ArgumentParameters: []*Parameter{signature.Parameters[0], signature.Parameters[0]},
		ArgumentTypes:      []*TypeSymbol{ScalarTypes.DynamicArrayOfString.Type(), ScalarTypes.DynamicArrayOfLong.Type()},
	}); got != ScalarTypes.DynamicArray.Type() {
		t.Fatalf("array_concat mixed-array return = %#v", got)
	}
}

func checkArrayIifFunction(t *testing.T, fn *FunctionSymbol, name string, hidden bool) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone || fn.IsHidden() != hidden {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s hidden:%t", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.IsHidden())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.CustomReturnType == nil || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	if got := signature.CustomReturnType(CustomReturnTypeContext{ArgumentTypes: []*TypeSymbol{ScalarTypes.DynamicArray.Type(), ScalarTypes.String.Type(), ScalarTypes.Long.Type()}}); got != ScalarTypes.DynamicArrayOfString {
		t.Fatalf("%s custom return = %#v", name, got)
	}
	bag := NewDynamicBagSymbol(NewColumnSymbol("id", ScalarTypes.String.Type()))
	array := NewDynamicArraySymbol(bag.Type())
	for testName, test := range map[string]struct {
		whenTrue *TypeSymbol
		want     Symbol
	}{
		"dynamic any":       {ScalarTypes.Dynamic.Type(), ScalarTypes.Dynamic.Type()},
		"dynamic primitive": {ScalarTypes.DynamicLong.Type(), ScalarTypes.DynamicLong.Type()},
		"dynamic bag":       {bag.Type(), bag.Type()},
		"dynamic array":     {array.Type(), array.Type()},
	} {
		got := signature.CustomReturnType(CustomReturnTypeContext{
			Signature:          signature,
			ArgumentParameters: signature.Parameters,
			ArgumentTypes:      []*TypeSymbol{ScalarTypes.DynamicArray.Type(), test.whenTrue, ScalarTypes.String.Type()},
		})
		if got != test.want {
			t.Fatalf("%s %s return = %#v, want %#v", name, testName, got, test.want)
		}
	}
}

func checkArrayIndexOfFunction(t *testing.T) {
	t.Helper()
	fn := Functions.ArrayIndexOf
	if fn.SymbolName() != "array_index_of" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("array_index_of metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Long.Type() || signature.MinArgumentCount != 2 || signature.MaxArgumentCount != 5 {
		t.Fatalf("array_index_of signature = %#v", signature)
	}
	for _, index := range []int{2, 3, 4} {
		if !signature.Parameters[index].IsOptional() {
			t.Fatalf("array_index_of parameter %d should be optional: %#v", index, signature.Parameters[index])
		}
	}
	if signature.Parameters[4].Name != "occurrence" || signature.Parameters[4].ArgumentKind != ArgumentKindConstant {
		t.Fatalf("array_index_of occurrence parameter = %#v", signature.Parameters[4])
	}
}

func checkArraySplitFunction(t *testing.T) {
	t.Helper()
	fn := Functions.ArraySplit
	if fn.SymbolName() != "array_split" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("array_split metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 2 {
		t.Fatalf("array_split signatures = %d", len(signatures))
	}
	if signatures[0].ReturnKind != ReturnTypeParameter0 || signatures[0].Parameters[1].TypeKind != ParameterTypeInteger {
		t.Fatalf("array_split index signature = %#v", signatures[0])
	}
	if signatures[1].ReturnKind != ReturnTypeParameter0 || signatures[1].Parameters[1].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("array_split indices signature = %#v", signatures[1])
	}
}

func checkReturnParameter0Function(t *testing.T, fn *FunctionSymbol, name string, parameterNames []string, parameterKinds []ParameterTypeKind, optionalLast bool) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeParameter0 || len(signature.Parameters) != len(parameterNames) {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	for i, parameterName := range parameterNames {
		parameter := signature.Parameters[i]
		if parameter.Name != parameterName || parameter.TypeKind != parameterKinds[i] {
			t.Fatalf("%s parameter %d = %#v", name, i, parameter)
		}
	}
	if optionalLast && !signature.Parameters[len(signature.Parameters)-1].IsOptional() {
		t.Fatalf("%s last parameter should be optional: %#v", name, signature.Parameters[len(signature.Parameters)-1])
	}
}

func checkArraySortFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.CustomReturnType == nil || signature.Layout != ParameterLayoutCustom || signature.Tabularity != TabularityScalar {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	if signature.MinArgumentCount != 1 || signature.MaxArgumentCount != 65 {
		t.Fatalf("%s argument count = %#v", name, signature)
	}
	if arrays := signature.Parameters[0]; arrays.Name != "arrays" || arrays.TypeKind != ParameterTypeDynamicArray || arrays.MinOccurring != 1 || arrays.MaxOccurring != 64 {
		t.Fatalf("%s arrays parameter = %#v", name, arrays)
	}
	if nullsLast := signature.Parameters[1]; nullsLast.Name != "nulls_last" || nullsLast.ArgumentKind != ArgumentKindConstant || !nullsLast.IsOptional() {
		t.Fatalf("%s nulls_last parameter = %#v", name, nullsLast)
	}
	arrays := signature.Parameters[0]
	nullsLast := signature.Parameters[1]
	parameters := signature.ArgumentParametersForTypes([]*TypeSymbol{
		ScalarTypes.DynamicArrayOfString.Type(),
		ScalarTypes.DynamicArrayOfLong.Type(),
		ScalarTypes.Bool.Type(),
	})
	if len(parameters) != 3 || parameters[0] != arrays || parameters[1] != arrays || parameters[2] != nullsLast {
		t.Fatalf("%s custom layout parameters = %#v", name, parameters)
	}
	nonFinalBool := signature.ArgumentParametersForTypes([]*TypeSymbol{
		ScalarTypes.DynamicArrayOfString.Type(),
		ScalarTypes.Bool.Type(),
		ScalarTypes.DynamicArrayOfLong.Type(),
	})
	if len(nonFinalBool) != 3 || nonFinalBool[0] != arrays || nonFinalBool[1] != arrays || nonFinalBool[2] != arrays {
		t.Fatalf("%s non-final bool parameters = %#v", name, nonFinalBool)
	}
	firstArgument := functionCallArgument(t, name+`(dynamic([1]), false)`, 0)
	boolArgument := functionCallArgument(t, name+`(dynamic([1]), false)`, 1)
	expressionParameters := signature.ArgumentParametersForExpressions(
		[]Expression{firstArgument, boolArgument},
		[]Symbol{ScalarTypes.DynamicArray.Type(), ScalarTypes.Unknown.Type()},
	)
	if len(expressionParameters) != 2 || expressionParameters[0] != arrays || expressionParameters[1] != nullsLast {
		t.Fatalf("%s literal bool custom layout parameters = %#v", name, expressionParameters)
	}
	columnBoolParameters := signature.ArgumentParametersForExpressions(
		[]Expression{firstArgument, boolArgument},
		[]Symbol{ScalarTypes.DynamicArray.Type(), NewColumnSymbol("Flag", ScalarTypes.Bool.Type())},
	)
	if len(columnBoolParameters) != 2 || columnBoolParameters[0] != arrays || columnBoolParameters[1] != nullsLast {
		t.Fatalf("%s column bool custom layout parameters = %#v", name, columnBoolParameters)
	}
	next := signature.NextPossibleParametersForTypes([]*TypeSymbol{ScalarTypes.DynamicArrayOfString.Type()})
	if len(next) != 1 || next[0] != arrays {
		t.Fatalf("%s custom next possible parameters = %#v", name, next)
	}
	result := signature.CustomReturnType(CustomReturnTypeContext{
		ArgumentTypes:       []*TypeSymbol{ScalarTypes.DynamicArrayOfString.Type(), ScalarTypes.DynamicArrayOfLong.Type(), ScalarTypes.Bool.Type()},
		ArgumentResultNames: []string{"Names", "Ids", "nulls_last"},
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("%s custom return = %T", name, result)
	}
	columns := tuple.Columns()
	if len(columns) != 2 || columns[0].SymbolName() != "Names_sorted" || columns[1].SymbolName() != "Ids_sorted" {
		t.Fatalf("%s custom columns = %#v", name, columns)
	}
	if columns[0].Type != ScalarTypes.DynamicArrayOfString.Type() || columns[1].Type != ScalarTypes.DynamicArrayOfLong.Type() {
		t.Fatalf("%s custom column types = %#v", name, columns)
	}
}

func checkDynamicArrayDeclaredFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, parameterNames []string, parameterKinds []ParameterTypeKind) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != returnType || len(signature.Parameters) != len(parameterNames) {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	for i, parameterName := range parameterNames {
		parameter := signature.Parameters[i]
		if parameter.Name != parameterName || parameter.TypeKind != parameterKinds[i] {
			t.Fatalf("%s parameter %d = %#v", name, i, parameter)
		}
	}
}

func checkBagPackFunction(t *testing.T, fn *FunctionSymbol, name string, obsolete bool) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone || fn.IsHidden() != obsolete {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s hidden:%t", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.IsHidden())
	}
	if obsolete && fn.Alternative() != "bag_pack" {
		t.Fatalf("%s alternative = %q", name, fn.Alternative())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.CustomReturnType == nil || signature.Layout != ParameterLayoutBlockRepeating {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	if signature.MinArgumentCount != 2 || signature.MaxArgumentCount != MaxRepeat*2 || !signature.IsValidArgumentCount(2) || signature.IsValidArgumentCount(3) || !signature.IsValidArgumentCount(4) {
		t.Fatalf("%s argument count/layout = %#v", name, signature)
	}
	if signature.Parameters[0].Name != "key" || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.String.Type() || signature.Parameters[1].Name != "value" || signature.Parameters[1].TypeKind != ParameterTypeScalar {
		t.Fatalf("%s parameters = %#v", name, signature.Parameters)
	}
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Signature:              signature,
		ArgumentParameters:     []*Parameter{signature.Parameters[0], signature.Parameters[1], signature.Parameters[0], signature.Parameters[1]},
		ArgumentTypes:          []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.Long.Type(), ScalarTypes.String.Type(), ScalarTypes.DateTime.Type()},
		ArgumentConstantValues: []any{"Account", nil, "Timestamp", nil},
	})
	bag, ok := result.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("%s custom return = %T", name, result)
	}
	properties := bag.Properties()
	if len(properties) != 2 || properties[0].SymbolName() != "Account" || properties[0].Type != ScalarTypes.Long.Type() || properties[1].SymbolName() != "Timestamp" || properties[1].Type != ScalarTypes.DateTime.Type() {
		t.Fatalf("%s custom properties = %#v", name, properties)
	}
}

func checkBagPackColumnsFunction(t *testing.T) {
	t.Helper()
	fn := Functions.BagPackColumns
	if fn.SymbolName() != "bag_pack_columns" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("bag_pack_columns metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	parameter := signature.Parameters[0]
	if signature.ReturnKind != ReturnTypeCustom || parameter.Name != "value" || parameter.TypeKind != ParameterTypeScalar || parameter.ArgumentKind != ArgumentKindColumn || parameter.MinOccurring != 1 || parameter.MaxOccurring != MaxRepeat {
		t.Fatalf("bag_pack_columns signature = %#v", signature)
	}
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	score := NewColumnSymbol("Score", ScalarTypes.Long.Type())
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Signature:          signature,
		ArgumentParameters: []*Parameter{parameter, parameter},
		ArgumentSemantics: []SemanticInfo{
			NewSemanticInfoForSymbol(account, account.Type),
			NewSemanticInfoForSymbol(score, score.Type),
		},
	})
	bag, ok := result.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("bag_pack_columns custom return = %T", result)
	}
	properties := bag.Properties()
	if len(properties) != 2 || properties[0] != account || properties[1] != score {
		t.Fatalf("bag_pack_columns properties = %#v", properties)
	}
}

func checkPackAllFunction(t *testing.T) {
	t.Helper()
	fn := Functions.PackAll
	if fn.SymbolName() != "pack_all" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("pack_all metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	parameter := signature.Parameters[0]
	if signature.ReturnKind != ReturnTypeCustom || parameter.Name != "ignore_null_empty" || parameter.TypeKind != ParameterTypeScalar || parameter.ArgumentKind != ArgumentKindLiteral || !parameter.IsOptional() {
		t.Fatalf("pack_all signature = %#v", signature)
	}
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	score := NewColumnSymbol("Score", ScalarTypes.Long.Type())
	result := signature.CustomReturnType(CustomReturnTypeContext{
		RowScope: NewTableSymbol("Events", account, score),
	})
	bag, ok := result.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("pack_all custom return = %T", result)
	}
	properties := bag.Properties()
	if len(properties) != 2 || properties[0] != account || properties[1] != score {
		t.Fatalf("pack_all properties = %#v", properties)
	}
}

func checkPackArrayFunction(t *testing.T) {
	t.Helper()
	fn := Functions.PackArray
	if fn.SymbolName() != "pack_array" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("pack_array metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 2 {
		t.Fatalf("pack_array signatures = %d", len(signatures))
	}
	first := signatures[0]
	if first.ReturnKind != ReturnTypeCustom || first.Parameters[0].Name != "value" || first.Parameters[0].TypeKind != ParameterTypeScalar || first.Parameters[0].MaxOccurring != MaxRepeat {
		t.Fatalf("pack_array value signature = %#v", first)
	}
	value := first.Parameters[0]
	result := first.CustomReturnType(CustomReturnTypeContext{
		Signature:          first,
		ArgumentParameters: []*Parameter{value, value},
		ArgumentTypes:      []*TypeSymbol{ScalarTypes.Int.Type(), ScalarTypes.Long.Type()},
	})
	if result != ScalarTypes.DynamicArrayOfLong {
		t.Fatalf("pack_array common return = %#v", result)
	}
	second := signatures[1]
	if second.ReturnKind != ReturnTypeCustom || second.Parameters[0].ArgumentKind != ArgumentKindStarOnly {
		t.Fatalf("pack_array star signature = %#v", second)
	}
	result = second.CustomReturnType(CustomReturnTypeContext{
		RowScope: NewTableSymbol("Events",
			NewColumnSymbol("Account", ScalarTypes.String.Type()),
			NewColumnSymbol("Name", ScalarTypes.String.Type())),
	})
	if result != ScalarTypes.DynamicArrayOfString {
		t.Fatalf("pack_array star return = %#v", result)
	}
}

func checkSetFunctions(t *testing.T) {
	t.Helper()
	checkDynamicArrayDeclaredFunction(t, Functions.SetHasElement, "set_has_element", ScalarTypes.Bool.Type(), []string{"set", "value"}, []ParameterTypeKind{ParameterTypeDynamicArray, ParameterTypeScalar})
	checkSetFunction(t, Functions.SetUnion, "set_union")
	checkSetFunction(t, Functions.SetIntersect, "set_intersect")
	checkSetFunction(t, Functions.SetDifference, "set_difference")
	checkDynamicArrayDeclaredFunction(t, Functions.SetEquals, "set_equals", ScalarTypes.Bool.Type(), []string{"set1", "set2"}, []ParameterTypeKind{ParameterTypeDynamicArray, ParameterTypeDynamicArray})
}

func checkSetFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	parameter := signature.Parameters[0]
	if signature.ReturnKind != ReturnTypeCustom || parameter.Name != "set" || parameter.TypeKind != ParameterTypeDynamicArray || parameter.MinOccurring != 2 || parameter.MaxOccurring != 64 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Signature:          signature,
		ArgumentParameters: []*Parameter{parameter, parameter},
		ArgumentTypes:      []*TypeSymbol{ScalarTypes.DynamicArrayOfLong.Type(), ScalarTypes.DynamicArrayOfLong.Type()},
	})
	if result != ScalarTypes.DynamicArrayOfLong.Type() {
		t.Fatalf("%s common return = %#v", name, result)
	}
}

func checkDynamicBagContinuationFunctions(t *testing.T) {
	t.Helper()
	checkBagMergeFunction(t)
	checkDynamicArrayDeclaredFunction(t, Functions.DynamicToJson, "dynamic_to_json", ScalarTypes.String.Type(), []string{"dynamic"}, []ParameterTypeKind{ParameterTypeScalar})
	checkBagRemoveKeysFunction(t)
	checkDynamicArrayDeclaredFunction(t, Functions.BagHasKey, "bag_has_key", ScalarTypes.Bool.Type(), []string{"bag", "key"}, []ParameterTypeKind{ParameterTypeDynamicBag, ParameterTypeDeclared})
	checkDynamicArrayDeclaredFunction(t, Functions.JaccardIndex, "jaccard_index", ScalarTypes.DynamicArrayOfReal.Type(), []string{"set"}, []ParameterTypeKind{ParameterTypeDynamicArray})
	if parameter := Functions.JaccardIndex.Signatures()[0].Parameters[0]; parameter.MinOccurring != 2 || parameter.MaxOccurring != 2 {
		t.Fatalf("jaccard_index set occurrence = %#v", parameter)
	}
	checkBagSetKeyFunction(t)
	checkDynamicArrayDeclaredFunction(t, Functions.BagZip, "bag_zip", ScalarTypes.DynamicBag.Type(), []string{"keys", "values"}, []ParameterTypeKind{ParameterTypeDynamicArray, ParameterTypeDynamicArray})
}

func checkBagMergeFunction(t *testing.T) {
	t.Helper()
	fn := Functions.BagMerge
	if fn.SymbolName() != "bag_merge" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("bag_merge metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	parameter := signature.Parameters[0]
	if signature.ReturnKind != ReturnTypeCustom || parameter.Name != "bag" || parameter.TypeKind != ParameterTypeDynamicBag || parameter.MinOccurring != 2 || parameter.MaxOccurring != 64 {
		t.Fatalf("bag_merge signature = %#v", signature)
	}
	left := ScalarTypes.GetDynamicBag(
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Shared", ScalarTypes.Int.Type()))
	right := ScalarTypes.GetDynamicBag(
		NewColumnSymbol("Shared", ScalarTypes.Long.Type()),
		NewColumnSymbol("Timestamp", ScalarTypes.DateTime.Type()))
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Signature:              signature,
		ArgumentParameters:     []*Parameter{parameter, parameter},
		ArgumentResultSymbols:  []Symbol{left, right},
		ArgumentConstantValues: []any{nil, nil},
	})
	bag, ok := result.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("bag_merge custom return = %T", result)
	}
	properties := bag.Properties()
	if len(properties) != 3 || properties[0].SymbolName() != "Account" || properties[1].SymbolName() != "Shared" || properties[2].SymbolName() != "Timestamp" {
		t.Fatalf("bag_merge properties = %#v", properties)
	}
	if properties[1].Type != ScalarTypes.Int.Type() {
		t.Fatalf("bag_merge common Shared type = %#v", properties[1].Type)
	}
}

func checkBagRemoveKeysFunction(t *testing.T) {
	t.Helper()
	fn := Functions.BagRemoveKeys
	if fn.SymbolName() != "bag_remove_keys" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("bag_remove_keys metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.Parameters[0].Name != "bag" || signature.Parameters[0].TypeKind != ParameterTypeDynamicBag || signature.Parameters[1].Name != "keys" || signature.Parameters[1].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("bag_remove_keys signature = %#v", signature)
	}
	source := ScalarTypes.GetDynamicBag(
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("RemoveMe", ScalarTypes.Long.Type()),
		NewColumnSymbol("Timestamp", ScalarTypes.DateTime.Type()))
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Signature:              signature,
		ArgumentParameters:     []*Parameter{signature.Parameters[0], signature.Parameters[1]},
		ArgumentResultSymbols:  []Symbol{source, ScalarTypes.DynamicArrayOfString},
		ArgumentConstantValues: []any{nil, []string{"RemoveMe"}},
	})
	bag, ok := result.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("bag_remove_keys custom return = %T", result)
	}
	properties := bag.Properties()
	if len(properties) != 2 || properties[0].SymbolName() != "Account" || properties[1].SymbolName() != "Timestamp" {
		t.Fatalf("bag_remove_keys properties = %#v", properties)
	}
}

func checkBagSetKeyFunction(t *testing.T) {
	t.Helper()
	fn := Functions.BagSetKey
	if fn.SymbolName() != "bag_set_key" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("bag_set_key metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.Parameters[0].TypeKind != ParameterTypeDynamicBag || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.String.Type() || signature.Parameters[1].ArgumentKind != ArgumentKindConstant || signature.Parameters[2].TypeKind != ParameterTypeScalar {
		t.Fatalf("bag_set_key signature = %#v", signature)
	}
	source := ScalarTypes.GetDynamicBag(NewColumnSymbol("Account", ScalarTypes.String.Type()))
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Signature:              signature,
		ArgumentParameters:     []*Parameter{signature.Parameters[0], signature.Parameters[1], signature.Parameters[2]},
		ArgumentResultSymbols:  []Symbol{source, ScalarTypes.String, ScalarTypes.Long},
		ArgumentConstantValues: []any{nil, "Score", nil},
	})
	bag, ok := result.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("bag_set_key custom return = %T", result)
	}
	properties := bag.Properties()
	if len(properties) != 2 || properties[0].SymbolName() != "Account" || properties[1].SymbolName() != "Score" || properties[1].Type != ScalarTypes.Long.Type() {
		t.Fatalf("bag_set_key added properties = %#v", properties)
	}

	nestedSource := ScalarTypes.GetDynamicBag(
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("Score", ScalarTypes.GetDynamicBag(NewColumnSymbol("Previous", ScalarTypes.Int.Type())).Type()))
	result = signature.CustomReturnType(CustomReturnTypeContext{
		Signature:              signature,
		ArgumentParameters:     []*Parameter{signature.Parameters[0], signature.Parameters[1], signature.Parameters[2]},
		ArgumentResultSymbols:  []Symbol{nestedSource, ScalarTypes.String, ScalarTypes.Long},
		ArgumentConstantValues: []any{nil, "$.Score.Value", nil},
	})
	bag, ok = result.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("bag_set_key JSON path return = %T", result)
	}
	scoreProperty, ok := bag.TryGetProperty("Score")
	if !ok {
		t.Fatalf("bag_set_key JSON path missing Score property = %#v", bag.Properties())
	}
	scoreBag, ok := DynamicBagForType(scoreProperty.Type)
	if !ok {
		t.Fatalf("bag_set_key JSON path Score type = %#v", scoreProperty.Type)
	}
	if previous, ok := scoreBag.TryGetProperty("Previous"); !ok || previous.Type != ScalarTypes.Int.Type() {
		t.Fatalf("bag_set_key JSON path preserved sibling = %#v", scoreBag.Properties())
	}
	if value, ok := scoreBag.TryGetProperty("Value"); !ok || value.Type != ScalarTypes.Long.Type() {
		t.Fatalf("bag_set_key JSON path nested value = %#v", scoreBag.Properties())
	}

	result = signature.CustomReturnType(CustomReturnTypeContext{
		Signature:              signature,
		ArgumentParameters:     []*Parameter{signature.Parameters[0], signature.Parameters[1], signature.Parameters[2]},
		ArgumentResultSymbols:  []Symbol{source, ScalarTypes.String, ScalarTypes.String},
		ArgumentConstantValues: []any{nil, "$[\"a b\"][\"c-d\"]", nil},
	})
	bag, ok = result.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("bag_set_key bracket JSON path return = %T", result)
	}
	spaceProperty, ok := bag.TryGetProperty("a b")
	if !ok {
		t.Fatalf("bag_set_key bracket JSON path missing property = %#v", bag.Properties())
	}
	spaceBag, ok := DynamicBagForType(spaceProperty.Type)
	if !ok {
		t.Fatalf("bag_set_key bracket JSON path property type = %#v", spaceProperty.Type)
	}
	if dash, ok := spaceBag.TryGetProperty("c-d"); !ok || dash.Type != ScalarTypes.String.Type() {
		t.Fatalf("bag_set_key bracket JSON path nested value = %#v", spaceBag.Properties())
	}

	result = signature.CustomReturnType(CustomReturnTypeContext{
		Signature:              signature,
		ArgumentParameters:     []*Parameter{signature.Parameters[0], signature.Parameters[1], signature.Parameters[2]},
		ArgumentResultSymbols:  []Symbol{source, ScalarTypes.String, ScalarTypes.Long},
		ArgumentConstantValues: []any{nil, "$[0].Score", nil},
	})
	if result != source {
		t.Fatalf("bag_set_key unsupported JSON path fallback = %#v", result)
	}
}

func checkPunycodeFunctions(t *testing.T) {
	t.Helper()
	checkDynamicArrayDeclaredFunction(t, Functions.PunycodeDecode, "punycode_to_string", ScalarTypes.String.Type(), []string{"string"}, []ParameterTypeKind{ParameterTypeDeclared})
	checkDynamicArrayDeclaredFunction(t, Functions.PunycodeEncode, "punycode_from_string", ScalarTypes.String.Type(), []string{"string"}, []ParameterTypeKind{ParameterTypeDeclared})
	checkDynamicArrayDeclaredFunction(t, Functions.PunycodeDomainDecode, "punycode_domain_from_string", ScalarTypes.String.Type(), []string{"string"}, []ParameterTypeKind{ParameterTypeDeclared})
	checkDynamicArrayDeclaredFunction(t, Functions.PunycodeDomainEncode, "punycode_domain_to_string", ScalarTypes.String.Type(), []string{"string"}, []ParameterTypeKind{ParameterTypeDeclared})
}

func checkDigestSeriesFunctions(t *testing.T) {
	t.Helper()
	checkPercentileTDigestFunction(t)
	checkPercentileArrayTDigestFunction(t)
	checkDigestRealFunction(t, Functions.PercentRankTDigest, "percentrank_tdigest", "digest")
	checkDigestRealFunction(t, Functions.RankTDigest, "rank_tdigest", "digest")
	checkDigestValidationFunction(t, Functions.TdigestIsValid, "tdigest_isvalid", "digest")
	checkDigestValidationFunction(t, Functions.HllIsValid, "hll_isvalid", "hll")
	checkDigestMergeFunction(t, Functions.TDigestMerge, "tdigest_merge", "tdigest", "tdigests_merge_result")
	checkDigestMergeFunction(t, Functions.MergeTDigest, "merge_tdigest", "tdigest", "tdigests_merge_result")
	checkDigestMergeFunction(t, Functions.HllMerge, "hll_merge", "hll", "hll_merged_result")
	checkDigestDCountFunction(t)
	checkHllNormalizeFunction(t)
	checkSeriesFirFunction(t)
	checkSeriesStatsFunction(t)
	checkSeriesStatsDynamicFunction(t)
	checkSeriesFftFunction(t, Functions.SeriesFft, "series_fft")
	checkSeriesFftFunction(t, Functions.SeriesIfft, "series_ifft")
	checkSeriesFitLineFunction(t)
	checkSeriesFitLineDynamicFunction(t)
	checkSeriesFit2LinesFunction(t)
	checkSeriesFit2LinesDynamicFunction(t)
	checkSeriesOutliersFunction(t)
	checkSeriesIIRFunction(t)
	checkSeriesPeriodsFunction(t, Functions.SeriesPeriodsDetect, "series_periods_detect", 4)
	checkSeriesPeriodsFunction(t, Functions.SeriesPeriodsValidate, "series_periods_validate", 17)
	checkSeriesFillFunctions(t)
	checkSeriesFitPolyFunction(t)
	checkSeriesBinaryFunctions(t)
	checkSeriesUnaryFunctions(t)
	checkSeriesSeasonalFunction(t)
	checkSeriesDecomposeFunctions(t)
	checkSeriesCorrelationFunctions(t)
	checkMathFunctions(t)
	checkGeoFunctions(t)
	checkGraphFunctions(t)
	checkIPAndFormattingFunctions(t)
	checkOtherFunctions(t)
}

func checkPercentileTDigestFunction(t *testing.T) {
	t.Helper()
	fn := Functions.PercentileTDigest
	if fn.SymbolName() != "percentile_tdigest" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNameAndFirstArgument {
		t.Fatalf("percentile_tdigest metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 2 {
		t.Fatalf("percentile_tdigest signatures = %d", len(signatures))
	}
	if signatures[0].DeclaredReturnType != ScalarTypes.Dynamic.Type() || signatures[0].Parameters[0].TypeKind != ParameterTypeDynamicArray || signatures[0].Parameters[1].DeclaredTypes[0] != ScalarTypes.Real.Type() {
		t.Fatalf("percentile_tdigest scalar signature = %#v", signatures[0])
	}
	if signatures[1].ReturnKind != ReturnTypeParameterNLiteral || signatures[1].Parameters[2].DeclaredTypes[0] != ScalarTypes.TypeLit.Type() || signatures[1].Parameters[2].ArgumentKind != ArgumentKindLiteral {
		t.Fatalf("percentile_tdigest typed signature = %#v", signatures[1])
	}
}

func checkPercentileArrayTDigestFunction(t *testing.T) {
	t.Helper()
	fn := Functions.PercentileArrayTDigest
	if fn.SymbolName() != "percentiles_array_tdigest" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNamePrefixAndFirstArgument || fn.ResultNamePrefix() != "percentile_tdigest" {
		t.Fatalf("percentiles_array_tdigest metadata = name:%s fold:%t result:%s prefix:%q", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.ResultNamePrefix())
	}
	signatures := fn.Signatures()
	if len(signatures) != 2 {
		t.Fatalf("percentiles_array_tdigest signatures = %d", len(signatures))
	}
	if signatures[0].DeclaredReturnType != ScalarTypes.DynamicArray.Type() || signatures[0].Parameters[1].Name != "percentile" || signatures[0].Parameters[1].MinOccurring != 1 || signatures[0].Parameters[1].MaxOccurring != MaxRepeat {
		t.Fatalf("percentiles_array_tdigest repeated percentile signature = %#v", signatures[0])
	}
	if signatures[1].DeclaredReturnType != ScalarTypes.DynamicArray.Type() || signatures[1].Parameters[1].Name != "percentiles" || signatures[1].Parameters[1].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("percentiles_array_tdigest array signature = %#v", signatures[1])
	}
}

func checkDigestRealFunction(t *testing.T, fn *FunctionSymbol, name string, digestParameter string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Real.Type() || signature.Parameters[0].Name != digestParameter || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.Parameters[1].Name != "value" || signature.Parameters[1].TypeKind != ParameterTypeScalar {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkDigestValidationFunction(t *testing.T, fn *FunctionSymbol, name string, digestParameter string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone || !fn.IsHidden() {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s hidden:%t", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.IsHidden())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Bool.Type() || signature.Parameters[0].Name != digestParameter || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.Parameters[1].Name != "value" || signature.Parameters[1].TypeKind != ParameterTypeScalar {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkDigestMergeFunction(t *testing.T, fn *FunctionSymbol, name string, parameterName string, prefix string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNamePrefixOnly || fn.ResultNamePrefix() != prefix {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s prefix:%q", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.ResultNamePrefix())
	}
	signature := fn.Signatures()[0]
	parameter := signature.Parameters[0]
	if signature.DeclaredReturnType != ScalarTypes.DynamicArray.Type() || parameter.Name != parameterName || parameter.TypeKind != ParameterTypeDynamicArray || parameter.MinOccurring != 2 || parameter.MaxOccurring != 16 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkDigestDCountFunction(t *testing.T) {
	t.Helper()
	fn := Functions.DCountHll
	if fn.SymbolName() != "dcount_hll" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNameAndFirstArgument {
		t.Fatalf("dcount_hll metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Long.Type() || signature.Parameters[0].Name != "hll" || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("dcount_hll signature = %#v", signature)
	}
}

func checkHllNormalizeFunction(t *testing.T) {
	t.Helper()
	fn := Functions.HllNormalize
	if fn.SymbolName() != "__hll_normalize" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument || !fn.IsHidden() {
		t.Fatalf("__hll_normalize metadata = name:%s fold:%t result:%s hidden:%t", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.IsHidden())
	}
	signature := fn.Signatures()[0]
	parameter := signature.Parameters[0]
	if signature.DeclaredReturnType != ScalarTypes.Dynamic.Type() || parameter.Name != "hll" || parameter.DeclaredTypes[0] != ScalarTypes.Dynamic.Type() || parameter.MinOccurring != 2 || parameter.MaxOccurring != 16 {
		t.Fatalf("__hll_normalize signature = %#v", signature)
	}
}

func checkSeriesFirFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesFir
	if fn.SymbolName() != "series_fir" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument {
		t.Fatalf("series_fir metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.DynamicArrayOfReal.Type() || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.Parameters[1].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("series_fir signature = %#v", signature)
	}
	if !signature.Parameters[2].IsOptional() || !signature.Parameters[3].IsOptional() || signature.Parameters[2].DeclaredTypes[0] != ScalarTypes.Bool.Type() || signature.Parameters[3].DeclaredTypes[0] != ScalarTypes.Bool.Type() {
		t.Fatalf("series_fir optional bool parameters = %#v", signature.Parameters)
	}
}

func checkSeriesStatsFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesStats
	if fn.SymbolName() != "series_stats" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("series_stats metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.Parameters[0].Name != "series" || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.Parameters[1].ArgumentKind != ArgumentKindConstant || !signature.Parameters[1].IsOptional() {
		t.Fatalf("series_stats signature = %#v", signature)
	}
	seriesArg := testNameReference("Score")
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments:           []Expression{seriesArg},
		ArgumentParameters:  []*Parameter{signature.Parameters[0]},
		ArgumentResultNames: []string{"Score"},
		Signature:           signature,
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("series_stats custom return = %T", result)
	}
	assertColumnNames(t, tuple.Columns(), []string{
		"series_stats_Score_min",
		"series_stats_Score_min_idx",
		"series_stats_Score_max",
		"series_stats_Score_max_idx",
		"series_stats_Score_avg",
		"series_stats_Score_stdev",
		"series_stats_Score_variance",
	})
	if tuple.Columns()[0].Type != ScalarTypes.Real.Type() || tuple.Columns()[1].Type != ScalarTypes.Long.Type() {
		t.Fatalf("series_stats column types = %#v", tuple.Columns())
	}
}

func checkSeriesStatsDynamicFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesStatsDynamic
	if fn.SymbolName() != "series_stats_dynamic" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("series_stats_dynamic metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	result := signature.CustomReturnType(CustomReturnTypeContext{Signature: signature})
	bag, ok := result.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("series_stats_dynamic custom return = %T", result)
	}
	assertColumnNames(t, bag.Properties(), []string{"min", "min_idx", "max", "max_idx", "avg", "stdev", "variance"})
}

func checkSeriesFftFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.Parameters[1].Name != "series_imaginary" || signature.Parameters[1].TypeKind != ParameterTypeDynamicArray || !signature.Parameters[1].IsOptional() {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	seriesArg := testNameReference("Signal")
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments:           []Expression{seriesArg},
		ArgumentParameters:  []*Parameter{signature.Parameters[0]},
		ArgumentResultNames: []string{"Signal"},
		Signature:           signature,
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("%s custom return = %T", name, result)
	}
	assertColumnNames(t, tuple.Columns(), []string{name + "_Signal_real", name + "_Signal_imag"})
	if tuple.Columns()[0].Type != ScalarTypes.DynamicArrayOfReal.Type() || tuple.Columns()[1].Type != ScalarTypes.DynamicArrayOfReal.Type() {
		t.Fatalf("%s custom column types = %#v", name, tuple.Columns())
	}
}

func checkSeriesFitLineFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesFitLine
	if fn.SymbolName() != "series_fit_line" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("series_fit_line metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.Parameters[0].Name != "series" || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("series_fit_line signature = %#v", signature)
	}
	seriesArg := testNameReference("Trend")
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments:           []Expression{seriesArg},
		ArgumentParameters:  []*Parameter{signature.Parameters[0]},
		ArgumentResultNames: []string{"Trend"},
		Signature:           signature,
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("series_fit_line custom return = %T", result)
	}
	assertColumnNames(t, tuple.Columns(), []string{
		"series_fit_line_Trend_rsquare",
		"series_fit_line_Trend_slope",
		"series_fit_line_Trend_variance",
		"series_fit_line_Trend_rvariance",
		"series_fit_line_Trend_interception",
		"series_fit_line_Trend_line_fit",
	})
	if tuple.Columns()[5].Type != ScalarTypes.DynamicArrayOfReal.Type() {
		t.Fatalf("series_fit_line line_fit type = %#v", tuple.Columns()[5].Type)
	}
}

func checkSeriesFitLineDynamicFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesFitLineDynamic
	if fn.SymbolName() != "series_fit_line_dynamic" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("series_fit_line_dynamic metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	result := signature.CustomReturnType(CustomReturnTypeContext{Signature: signature})
	bag, ok := result.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("series_fit_line_dynamic custom return = %T", result)
	}
	assertColumnNames(t, bag.Properties(), []string{"rsquare", "slope", "variance", "rvariance", "interception", "line_fit"})
	if bag.Properties()[5].Type != ScalarTypes.DynamicArrayOfReal.Type() {
		t.Fatalf("series_fit_line_dynamic line_fit type = %#v", bag.Properties()[5].Type)
	}
}

func checkSeriesFit2LinesFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesFit2Lines
	if fn.SymbolName() != "series_fit_2lines" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("series_fit_2lines metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.Parameters[0].Name != "array" || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("series_fit_2lines signature = %#v", signature)
	}
	arrayArg := testNameReference("Trend")
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments:           []Expression{arrayArg},
		ArgumentParameters:  []*Parameter{signature.Parameters[0]},
		ArgumentResultNames: []string{"Trend"},
		Signature:           signature,
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("series_fit_2lines custom return = %T", result)
	}
	assertColumnNames(t, tuple.Columns(), []string{
		"series_fit_2lines_Trend_rsquare",
		"series_fit_2lines_Trend_split_idx",
		"series_fit_2lines_Trend_variance",
		"series_fit_2lines_Trend_rvariance",
		"series_fit_2lines_Trend_line_fit",
		"series_fit_2lines_Trend_right_rsquare",
		"series_fit_2lines_Trend_right_slope",
		"series_fit_2lines_Trend_right_interception",
		"series_fit_2lines_Trend_right_variance",
		"series_fit_2lines_Trend_right_rvariance",
		"series_fit_2lines_Trend_left_rsquare",
		"series_fit_2lines_Trend_left_slope",
		"series_fit_2lines_Trend_left_interception",
		"series_fit_2lines_Trend_left_variance",
		"series_fit_2lines_Trend_left_rvariance",
	})
	if tuple.Columns()[1].Type != ScalarTypes.Long.Type() || tuple.Columns()[4].Type != ScalarTypes.DynamicArrayOfReal.Type() {
		t.Fatalf("series_fit_2lines column types = %#v", tuple.Columns())
	}
}

func checkSeriesFit2LinesDynamicFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesFit2LinesDynamic
	if fn.SymbolName() != "series_fit_2lines_dynamic" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("series_fit_2lines_dynamic metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.Parameters[0].Name != "series" || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("series_fit_2lines_dynamic signature = %#v", signature)
	}
	result := signature.CustomReturnType(CustomReturnTypeContext{Signature: signature})
	bag, ok := result.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("series_fit_2lines_dynamic custom return = %T", result)
	}
	assertColumnNames(t, bag.Properties(), []string{"rsquare", "split_idx", "variance", "rvariance", "line_fit", "right_rsquare", "right_slope", "right_interception", "right_variance", "right_rvariance", "left_rsquare", "left_slope", "left_interception", "left_variance", "left_rvariance"})
}

func checkSeriesOutliersFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesOutliers
	if fn.SymbolName() != "series_outliers" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument {
		t.Fatalf("series_outliers metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 2 {
		t.Fatalf("series_outliers signatures = %d", len(signatures))
	}
	if signatures[0].ReturnKind != ReturnTypeParameter0 || signatures[0].MinArgumentCount != 1 || signatures[0].MaxArgumentCount != 3 || signatures[0].Parameters[1].DeclaredTypes[0] != ScalarTypes.String.Type() || signatures[0].Parameters[2].TypeKind != ParameterTypeNumber {
		t.Fatalf("series_outliers short signature = %#v", signatures[0])
	}
	if signatures[1].ReturnKind != ReturnTypeParameter0 || signatures[1].MinArgumentCount != 4 || signatures[1].MaxArgumentCount != 6 || signatures[1].Parameters[4].DeclaredTypes[0] != ScalarTypes.Real.Type() || signatures[1].Parameters[5].TypeKind != ParameterTypeInteger {
		t.Fatalf("series_outliers percentile signature = %#v", signatures[1])
	}
}

func checkSeriesIIRFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesIIR
	if fn.SymbolName() != "series_iir" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument {
		t.Fatalf("series_iir metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.DynamicArrayOfReal.Type() || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.Parameters[1].Name != "numerators" || signature.Parameters[2].Name != "denominators" {
		t.Fatalf("series_iir signature = %#v", signature)
	}
}

func checkSeriesPeriodsFunction(t *testing.T, fn *FunctionSymbol, name string, maxArguments int) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.MaxArgumentCount != maxArguments {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	if name == "series_periods_validate" {
		if signature.Parameters[1].Name != "period" || signature.Parameters[1].TypeKind != ParameterTypeNumber || signature.Parameters[1].MinOccurring != 1 || signature.Parameters[1].MaxOccurring != 16 {
			t.Fatalf("%s period parameter = %#v", name, signature.Parameters[1])
		}
	}
	seriesArg := testNameReference("Signal")
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments:           []Expression{seriesArg},
		ArgumentParameters:  []*Parameter{signature.Parameters[0]},
		ArgumentResultNames: []string{"Signal"},
		Signature:           signature,
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("%s custom return = %T", name, result)
	}
	assertColumnNames(t, tuple.Columns(), []string{name + "_Signal_periods", name + "_Signal_scores"})
}

func checkSeriesFillFunctions(t *testing.T) {
	t.Helper()
	checkSeriesFillFunction(t, Functions.SeriesFillBackwards, "series_fill_backward", 1, 2)
	checkSeriesFillFunction(t, Functions.SeriesFillForward, "series_fill_forward", 1, 2)
	checkSeriesFillFunction(t, Functions.SeriesFillConst, "series_fill_const", 1, 3)
	checkSeriesFillFunction(t, Functions.SeriesFillLinear, "series_fill_linear", 1, 4)
	if parameter := Functions.SeriesFillLinear.Signatures()[0].Parameters[2]; parameter.Name != "fill_edges" || parameter.DeclaredTypes[0] != ScalarTypes.Bool.Type() || !parameter.IsOptional() {
		t.Fatalf("series_fill_linear fill_edges parameter = %#v", parameter)
	}
}

func checkSeriesFillFunction(t *testing.T, fn *FunctionSymbol, name string, minArguments int, maxArguments int) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeParameter0 || signature.MinArgumentCount != minArguments || signature.MaxArgumentCount != maxArguments || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkSeriesFitPolyFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesFitPoly
	if fn.SymbolName() != "series_fit_poly" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("series_fit_poly metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.Parameters[0].Name != "y_series" || signature.Parameters[1].Name != "x_series" || !signature.Parameters[1].IsOptional() || signature.Parameters[2].DeclaredTypes[0] != ScalarTypes.Int.Type() || !signature.Parameters[2].IsOptional() {
		t.Fatalf("series_fit_poly signature = %#v", signature)
	}
	yArg := testNameReference("Y")
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments:           []Expression{yArg},
		ArgumentParameters:  []*Parameter{signature.Parameters[0]},
		ArgumentResultNames: []string{"Y"},
		Signature:           signature,
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("series_fit_poly custom return = %T", result)
	}
	assertColumnNames(t, tuple.Columns(), []string{"series_fit_poly_Y_rsquare", "series_fit_poly_Y_coefficients", "series_fit_poly_Y_variance", "series_fit_poly_Y_rvariance", "series_fit_poly_Y_poly_fit"})
	if tuple.Columns()[1].Type != ScalarTypes.DynamicArrayOfReal.Type() || tuple.Columns()[4].Type != ScalarTypes.DynamicArrayOfReal.Type() {
		t.Fatalf("series_fit_poly dynamic array column types = %#v", tuple.Columns())
	}
}

func checkSeriesBinaryFunctions(t *testing.T) {
	t.Helper()
	checkSeriesBinaryFunction(t, Functions.SeriesAdd, "series_add", ScalarTypes.DynamicArray.Type())
	checkSeriesBinaryFunction(t, Functions.SeriesSubtract, "series_subtract", ScalarTypes.DynamicArray.Type())
	checkSeriesBinaryFunction(t, Functions.SeriesMultiply, "series_multiply", ScalarTypes.DynamicArray.Type())
	checkSeriesBinaryFunction(t, Functions.SeriesDivide, "series_divide", ScalarTypes.DynamicArrayOfReal.Type())
	checkSeriesBinaryFunction(t, Functions.SeriesPow, "series_pow", ScalarTypes.DynamicArrayOfReal.Type())
	checkSeriesBinaryFunction(t, Functions.SeriesGreater, "series_greater", ScalarTypes.DynamicArrayOfBool.Type())
	checkSeriesBinaryFunction(t, Functions.SeriesGreaterEquals, "series_greater_equals", ScalarTypes.DynamicArrayOfBool.Type())
	checkSeriesBinaryFunction(t, Functions.SeriesLess, "series_less", ScalarTypes.DynamicArrayOfBool.Type())
	checkSeriesBinaryFunction(t, Functions.SeriesLessEquals, "series_less_equals", ScalarTypes.DynamicArrayOfBool.Type())
	checkSeriesBinaryFunction(t, Functions.SeriesEquals, "series_equals", ScalarTypes.DynamicArrayOfBool.Type())
	checkSeriesBinaryFunction(t, Functions.SeriesNotEquals, "series_not_equals", ScalarTypes.DynamicArrayOfBool.Type())
}

func checkSeriesBinaryFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 3 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	for i, signature := range signatures {
		if signature.DeclaredReturnType != returnType || signature.MinArgumentCount != 2 || signature.MaxArgumentCount != 2 {
			t.Fatalf("%s signature %d = %#v", name, i, signature)
		}
	}
	if signatures[0].Parameters[0].TypeKind != ParameterTypeDynamicArray || signatures[0].Parameters[1].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("%s array-array signature = %#v", name, signatures[0])
	}
	if signatures[1].Parameters[0].TypeKind != ParameterTypeDynamicArray || signatures[1].Parameters[1].TypeKind != ParameterTypeNumber {
		t.Fatalf("%s array-number signature = %#v", name, signatures[1])
	}
	if signatures[2].Parameters[0].TypeKind != ParameterTypeNumber || signatures[2].Parameters[1].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("%s number-array signature = %#v", name, signatures[2])
	}
}

func checkSeriesUnaryFunctions(t *testing.T) {
	t.Helper()
	checkSeriesUnaryFunction(t, Functions.SeriesExp, "series_exp", ScalarTypes.DynamicArrayOfReal.Type())
	checkSeriesUnaryFunction(t, Functions.SeriesSign, "series_sign", ScalarTypes.DynamicArrayOfLong.Type())
	checkSeriesUnaryReturnKindFunction(t, Functions.SeriesAbs, "series_abs", ReturnTypeParameter0)
	checkSeriesUnaryFunction(t, Functions.SeriesSin, "series_sin", ScalarTypes.DynamicArrayOfReal.Type())
	checkSeriesUnaryFunction(t, Functions.SeriesAsin, "series_asin", ScalarTypes.DynamicArrayOfReal.Type())
	checkSeriesUnaryFunction(t, Functions.SeriesCos, "series_cos", ScalarTypes.DynamicArrayOfReal.Type())
	checkSeriesUnaryFunction(t, Functions.SeriesAcos, "series_acos", ScalarTypes.DynamicArrayOfReal.Type())
	checkSeriesUnaryFunction(t, Functions.SeriesTan, "series_tan", ScalarTypes.DynamicArrayOfReal.Type())
	checkSeriesUnaryFunction(t, Functions.SeriesAtan, "series_atan", ScalarTypes.DynamicArrayOfReal.Type())
	checkSeriesUnaryFunction(t, Functions.SeriesMagnitude, "series_magnitude", ScalarTypes.Real.Type())
	checkSeriesUnaryFunction(t, Functions.SeriesSum, "series_sum", ScalarTypes.Real.Type())
	checkSeriesUnaryFunction(t, Functions.SeriesProduct, "series_product", ScalarTypes.Real.Type())
	checkSeriesUnaryFunction(t, Functions.SeriesLog, "series_log", ScalarTypes.DynamicArrayOfReal.Type())
	checkSeriesUnaryFunction(t, Functions.SeriesFloor, "series_floor", ScalarTypes.DynamicArray.Type())
	checkSeriesUnaryReturnKindFunction(t, Functions.SeriesCeiling, "series_ceiling", ReturnTypeParameter0)
	checkSeriesUnaryFunction(t, Functions.ArraySum, "array_sum", ScalarTypes.Real.Type())
	if !Functions.ArraySum.IsHidden() || Functions.ArraySum.Alternative() != "series_sum" {
		t.Fatalf("array_sum obsolete metadata = hidden:%t alt:%q", Functions.ArraySum.IsHidden(), Functions.ArraySum.Alternative())
	}
}

func checkSeriesUnaryFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != returnType || signature.Parameters[0].Name != "series" || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.MinArgumentCount != 1 || signature.MaxArgumentCount != 1 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkSeriesUnaryReturnKindFunction(t *testing.T, fn *FunctionSymbol, name string, returnKind ReturnTypeKind) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != returnKind || signature.Parameters[0].Name != "series" || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.MinArgumentCount != 1 || signature.MaxArgumentCount != 1 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkSeriesSeasonalFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesSeasonal
	if fn.SymbolName() != "series_seasonal" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("series_seasonal metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 3 {
		t.Fatalf("series_seasonal signatures = %d", len(signatures))
	}
	for i, signature := range signatures {
		if signature.DeclaredReturnType != ScalarTypes.DynamicArrayOfReal.Type() || signature.Parameters[0].Name != "series" || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.MinArgumentCount != i+1 || signature.MaxArgumentCount != i+1 {
			t.Fatalf("series_seasonal signature %d = %#v", i, signature)
		}
	}
	if signatures[1].Parameters[1].TypeKind != ParameterTypeInteger || signatures[2].Parameters[2].TypeKind != ParameterTypeInteger {
		t.Fatalf("series_seasonal integer parameters = %#v / %#v", signatures[1], signatures[2])
	}
}

func checkSeriesDecomposeFunctions(t *testing.T) {
	t.Helper()
	checkSeriesDecomposeFunction(t, Functions.SeriesDecompose, "series_decompose", 1, 5)
	checkSeriesDecomposeFunction(t, Functions.SeriesDecomposeForecast, "series_decompose_forecast", 2, 5)
	checkSeriesDecomposeAnomaliesFunction(t)
}

func checkSeriesDecomposeFunction(t *testing.T, fn *FunctionSymbol, name string, minArguments int, maxArguments int) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.MinArgumentCount != minArguments || signature.MaxArgumentCount != maxArguments || signature.Parameters[0].Name != "series" || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	if name == "series_decompose" {
		if signature.Parameters[1].Name != "period" || signature.Parameters[1].TypeKind != ParameterTypeInteger || !signature.Parameters[1].IsOptional() || signature.Parameters[2].DeclaredTypes[0] != ScalarTypes.String.Type() || !signature.Parameters[4].IsOptional() {
			t.Fatalf("%s optional parameters = %#v", name, signature.Parameters)
		}
	} else if signature.Parameters[1].Name != "test_points" || signature.Parameters[1].TypeKind != ParameterTypeInteger || signature.Parameters[1].IsOptional() || signature.Parameters[2].Name != "period" || !signature.Parameters[2].IsOptional() {
		t.Fatalf("%s forecast parameters = %#v", name, signature.Parameters)
	}
	seriesArg := testNameReference("Metric")
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments:           []Expression{seriesArg},
		ArgumentParameters:  []*Parameter{signature.Parameters[0]},
		ArgumentResultNames: []string{"Metric"},
		Signature:           signature,
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("%s custom return = %T", name, result)
	}
	assertColumnNames(t, tuple.Columns(), []string{
		name + "_Metric_baseline",
		name + "_Metric_seasonal",
		name + "_Metric_trend",
		name + "_Metric_residual",
	})
	if tuple.Columns()[0].Type != ScalarTypes.DynamicArray.Type() || tuple.Columns()[3].Type != ScalarTypes.DynamicArray.Type() {
		t.Fatalf("%s custom column types = %#v", name, tuple.Columns())
	}
}

func checkSeriesDecomposeAnomaliesFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesDecomposeAnomalies
	if fn.SymbolName() != "series_decompose_anomalies" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("series_decompose_anomalies metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.MinArgumentCount != 1 || signature.MaxArgumentCount != 7 || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.Parameters[1].TypeKind != ParameterTypeNumber || signature.Parameters[2].TypeKind != ParameterTypeInteger || signature.Parameters[5].DeclaredTypes[0] != ScalarTypes.String.Type() {
		t.Fatalf("series_decompose_anomalies signature = %#v", signature)
	}
	for i := 1; i < len(signature.Parameters); i++ {
		if !signature.Parameters[i].IsOptional() {
			t.Fatalf("series_decompose_anomalies parameter %d is not optional: %#v", i, signature.Parameters[i])
		}
	}
	seriesArg := testNameReference("Metric")
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments:           []Expression{seriesArg},
		ArgumentParameters:  []*Parameter{signature.Parameters[0]},
		ArgumentResultNames: []string{"Metric"},
		Signature:           signature,
	})
	tuple, ok := result.(*TupleSymbol)
	if !ok {
		t.Fatalf("series_decompose_anomalies custom return = %T", result)
	}
	assertColumnNames(t, tuple.Columns(), []string{
		"series_decompose_anomalies_Metric_ad_flag",
		"series_decompose_anomalies_Metric_ad_score",
		"series_decompose_anomalies_Metric_baseline",
	})
	if tuple.Columns()[0].Type != ScalarTypes.DynamicArray.Type() || tuple.Columns()[2].Type != ScalarTypes.DynamicArray.Type() {
		t.Fatalf("series_decompose_anomalies custom column types = %#v", tuple.Columns())
	}
}

func checkSeriesCorrelationFunctions(t *testing.T) {
	t.Helper()
	checkTwoSeriesRealFunction(t, Functions.SeriesPearsonCorrelation, "series_pearson_correlation")
	checkSeriesDotProductFunction(t)
	checkSeriesCosineSimilarityFunction(t)
}

func checkTwoSeriesRealFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Real.Type() || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.Parameters[1].TypeKind != ParameterTypeDynamicArray || signature.MinArgumentCount != 2 || signature.MaxArgumentCount != 2 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkSeriesDotProductFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesDotProduct
	if fn.SymbolName() != "series_dot_product" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("series_dot_product metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 3 {
		t.Fatalf("series_dot_product signatures = %d", len(signatures))
	}
	for i, signature := range signatures {
		if signature.DeclaredReturnType != ScalarTypes.Real.Type() || signature.MinArgumentCount != 2 || signature.MaxArgumentCount != 2 {
			t.Fatalf("series_dot_product signature %d = %#v", i, signature)
		}
	}
	if signatures[0].Parameters[0].TypeKind != ParameterTypeDynamicArray || signatures[0].Parameters[1].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("series_dot_product array-array signature = %#v", signatures[0])
	}
	if signatures[1].Parameters[0].TypeKind != ParameterTypeDynamicArray || signatures[1].Parameters[1].TypeKind != ParameterTypeNumber {
		t.Fatalf("series_dot_product array-number signature = %#v", signatures[1])
	}
	if signatures[2].Parameters[0].TypeKind != ParameterTypeNumber || signatures[2].Parameters[1].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("series_dot_product number-array signature = %#v", signatures[2])
	}
}

func checkSeriesCosineSimilarityFunction(t *testing.T) {
	t.Helper()
	fn := Functions.SeriesCosineSimilarity
	if fn.SymbolName() != "series_cosine_similarity" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("series_cosine_similarity metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Real.Type() || signature.MinArgumentCount != 2 || signature.MaxArgumentCount != 4 || signature.Parameters[0].TypeKind != ParameterTypeDynamicArray || signature.Parameters[1].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("series_cosine_similarity signature = %#v", signature)
	}
	if signature.Parameters[2].DeclaredTypes[0] != ScalarTypes.Real.Type() || !signature.Parameters[2].IsOptional() || signature.Parameters[3].DeclaredTypes[0] != ScalarTypes.Real.Type() || !signature.Parameters[3].IsOptional() {
		t.Fatalf("series_cosine_similarity optional magnitudes = %#v", signature.Parameters)
	}
}

func checkMathFunctions(t *testing.T) {
	t.Helper()
	checkRoundFunction(t)
	checkCeilingFunction(t)
	checkMathRealFirstArgumentFunction(t, Functions.Pow, "pow", []string{"base", "exponent"})
	checkMathRealFirstArgumentFunction(t, Functions.Sqrt, "sqrt", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Log, "log", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Log2, "log2", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Log10, "log10", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Exp, "exp", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Exp2, "exp2", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Exp10, "exp10", []string{"number"})
	checkPiFunction(t)
	checkMathRealFirstArgumentFunction(t, Functions.Cos, "cos", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Sin, "sin", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Tan, "tan", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Acos, "acos", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Asin, "asin", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Atan, "atan", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Atan2, "atan2", []string{"y", "x"})
	checkAbsFunction(t)
	checkMathRealNoneFunction(t, Functions.Cot, "cot", []string{"number"})
	checkMathRealNoneFunction(t, Functions.Degrees, "degrees", []string{"radians"})
	checkMathRealNoneFunction(t, Functions.Radians, "radians", []string{"degrees"})
	checkMathRealNoneFunction(t, Functions.Sign, "sign", []string{"number"})
	checkRandFunction(t)
	checkMathRealNoneFunction(t, Functions.BetaCdf, "beta_cdf", []string{"x", "alpha", "beta"})
	checkMathRealNoneFunction(t, Functions.BetaInv, "beta_inv", []string{"probability", "alpha", "beta"})
	checkMathRealNoneFunction(t, Functions.BetaPdf, "beta_pdf", []string{"x", "alpha", "beta"})
	checkMathRealFirstArgumentFunction(t, Functions.Gamma, "gamma", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.LogGamma, "loggamma", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Erf, "erf", []string{"number"})
	checkMathRealFirstArgumentFunction(t, Functions.Erfc, "erfc", []string{"number"})
	checkMathBoolRealFunction(t, Functions.IsNan, "isnan")
	checkMathBoolRealFunction(t, Functions.IsInf, "isinf")
	checkMathBoolRealFunction(t, Functions.IsFinite, "isfinite")
	checkCommonVariadicFunction(t, Functions.Coalesce, "coalesce", ParameterTypeCommonScalar)
	checkCommonVariadicFunction(t, Functions.MaxOf, "max_of", ParameterTypeCommonOrderable)
	checkCommonVariadicFunction(t, Functions.MinOf, "min_of", ParameterTypeCommonOrderable)
	checkWelchTestFunction(t)
}

func checkRoundFunction(t *testing.T) {
	t.Helper()
	fn := Functions.Round
	if fn.SymbolName() != "round" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument {
		t.Fatalf("round metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeParameter0 || signature.MinArgumentCount != 1 || signature.MaxArgumentCount != 2 || signature.Parameters[0].Name != "number" || signature.Parameters[0].TypeKind != ParameterTypeNumber {
		t.Fatalf("round signature = %#v", signature)
	}
	if signature.Parameters[1].Name != "precision" || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.Long.Type() || signature.Parameters[1].ArgumentKind != ArgumentKindConstant || !signature.Parameters[1].IsOptional() {
		t.Fatalf("round precision parameter = %#v", signature.Parameters[1])
	}
}

func checkCeilingFunction(t *testing.T) {
	t.Helper()
	fn := Functions.Ceiling
	if fn.SymbolName() != "ceiling" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument {
		t.Fatalf("ceiling metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeParameter0 || signature.Parameters[0].Name != "number" || signature.Parameters[0].TypeKind != ParameterTypeNumber || signature.MinArgumentCount != 1 || signature.MaxArgumentCount != 1 {
		t.Fatalf("ceiling signature = %#v", signature)
	}
}

func checkMathRealFirstArgumentFunction(t *testing.T, fn *FunctionSymbol, name string, parameterNames []string) {
	t.Helper()
	checkMathRealFunction(t, fn, name, ResultNameFirstArgument, true, parameterNames)
}

func checkMathRealNoneFunction(t *testing.T, fn *FunctionSymbol, name string, parameterNames []string) {
	t.Helper()
	checkMathRealFunction(t, fn, name, ResultNameNone, true, parameterNames)
}

func checkMathRealFunction(t *testing.T, fn *FunctionSymbol, name string, resultNameKind ResultNameKind, foldable bool, parameterNames []string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() != foldable || fn.ResultNameKind() != resultNameKind {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Real.Type() || signature.MinArgumentCount != len(parameterNames) || signature.MaxArgumentCount != len(parameterNames) {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	for i, parameterName := range parameterNames {
		if signature.Parameters[i].Name != parameterName || signature.Parameters[i].TypeKind != ParameterTypeNumber {
			t.Fatalf("%s parameter %d = %#v", name, i, signature.Parameters[i])
		}
	}
}

func checkPiFunction(t *testing.T) {
	t.Helper()
	fn := Functions.PI
	if fn.SymbolName() != "pi" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("pi metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Real.Type() || signature.MinArgumentCount != 0 || signature.MaxArgumentCount != 0 {
		t.Fatalf("pi signature = %#v", signature)
	}
}

func checkAbsFunction(t *testing.T) {
	t.Helper()
	fn := Functions.Abs
	if fn.SymbolName() != "abs" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument {
		t.Fatalf("abs metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 3 {
		t.Fatalf("abs signatures = %d", len(signatures))
	}
	if signatures[0].DeclaredReturnType != ScalarTypes.Long.Type() || signatures[0].Parameters[0].TypeKind != ParameterTypeInteger {
		t.Fatalf("abs integer signature = %#v", signatures[0])
	}
	if signatures[1].ReturnKind != ReturnTypeParameter0 || signatures[1].Parameters[0].TypeKind != ParameterTypeNumber {
		t.Fatalf("abs number signature = %#v", signatures[1])
	}
	if signatures[2].DeclaredReturnType != ScalarTypes.TimeSpan.Type() || signatures[2].Parameters[0].DeclaredTypes[0] != ScalarTypes.TimeSpan.Type() {
		t.Fatalf("abs timespan signature = %#v", signatures[2])
	}
}

func checkRandFunction(t *testing.T) {
	t.Helper()
	fn := Functions.Rand
	if fn.SymbolName() != "rand" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument {
		t.Fatalf("rand metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Real.Type() || signature.MinArgumentCount != 0 || signature.MaxArgumentCount != 1 || signature.Parameters[0].TypeKind != ParameterTypeInteger || !signature.Parameters[0].IsOptional() {
		t.Fatalf("rand signature = %#v", signature)
	}
}

func checkMathBoolRealFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Bool.Type() || signature.Parameters[0].Name != "number" || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.Real.Type() {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkCommonVariadicFunction(t *testing.T, fn *FunctionSymbol, name string, parameterType ParameterTypeKind) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	parameter := signature.Parameters[0]
	if signature.ReturnKind != ReturnTypeCommon || signature.MinArgumentCount != 2 || signature.MaxArgumentCount != 64 || parameter.Name != "arg" || parameter.TypeKind != parameterType || parameter.MinOccurring != 2 || parameter.MaxOccurring != 64 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkWelchTestFunction(t *testing.T) {
	t.Helper()
	fn := Functions.WelchTest
	if fn.SymbolName() != "welch_test" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("welch_test metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	want := []string{"mean1", "variance1", "count1", "mean2", "variance2", "count2"}
	if signature.DeclaredReturnType != ScalarTypes.Real.Type() || signature.MinArgumentCount != len(want) || signature.MaxArgumentCount != len(want) {
		t.Fatalf("welch_test signature = %#v", signature)
	}
	for i, parameterName := range want {
		if signature.Parameters[i].Name != parameterName || signature.Parameters[i].TypeKind != ParameterTypeNumber {
			t.Fatalf("welch_test parameter %d = %#v", i, signature.Parameters[i])
		}
	}
}

func checkGraphFunctions(t *testing.T) {
	t.Helper()
	checkGraphBoolFunction(t, Functions.GraphAll, "all")
	checkGraphBoolFunction(t, Functions.Any, "any")
	checkGraphMapFunction(t)
	checkGraphInnerNodesFunction(t)
	checkGraphScalarFunction(t, Functions.NodeId, "node_id", ScalarTypes.String.Type(), "node")
	checkGraphScalarFunction(t, Functions.NodeDegreeIn, "node_degree_in", ScalarTypes.Long.Type(), "node")
	checkGraphScalarFunction(t, Functions.NodeDegreeOut, "node_degree_out", ScalarTypes.Long.Type(), "node")
	checkGraphScalarFunction(t, Functions.Labels, "labels", ScalarTypes.DynamicArrayOfString.Type(), "pattern_element")
}

func checkGraphBoolFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	checkGraphScalarFunction(t, fn, name, ScalarTypes.Bool.Type(), "pattern_element")
	signature := fn.Signatures()[0]
	if len(signature.Parameters) != 2 || signature.Parameters[1].Name != "expression" || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.Bool.Type() || signature.Parameters[1].ArgumentKind != ArgumentKindExpressionParameter0Element {
		t.Fatalf("%s expression parameter = %#v", name, signature.Parameters)
	}
}

func checkGraphScalarFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, firstParameter string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.IsHidden() || fn.ResultNameKind() != ResultNameNone || fn.CustomAvailability() == nil {
		t.Fatalf("%s metadata = name:%s fold:%t hidden:%t result:%s availability:%t", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.IsHidden(), fn.ResultNameKind(), fn.CustomAvailability() != nil)
	}
	if fn.IsAvailable(CustomAvailabilityContext{}) || !fn.IsAvailable(CustomAvailabilityContext{InGraphWhereOrProjectClause: true}) {
		t.Fatalf("%s availability did not match graph where/project scope", name)
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != returnType || len(signature.Parameters) == 0 || signature.Parameters[0].Name != firstParameter || signature.Parameters[0].TypeKind != ParameterTypeScalar {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	if len(signature.Parameters) == 1 && (name == "node_id" || name == "node_degree_in" || name == "node_degree_out" || name == "labels") && !signature.Parameters[0].IsOptional() {
		t.Fatalf("%s optional parameter = %#v", name, signature.Parameters[0])
	}
}

func checkGraphMapFunction(t *testing.T) {
	t.Helper()
	fn := Functions.Map
	if fn.SymbolName() != "map" || fn.IsConstantFoldable() || fn.IsHidden() || fn.ResultNameKind() != ResultNameNone || fn.CustomAvailability() == nil {
		t.Fatalf("map metadata = name:%s fold:%t hidden:%t result:%s availability:%t", fn.SymbolName(), fn.IsConstantFoldable(), fn.IsHidden(), fn.ResultNameKind(), fn.CustomAvailability() != nil)
	}
	if fn.IsAvailable(CustomAvailabilityContext{}) || !fn.IsAvailable(CustomAvailabilityContext{InGraphWhereOrProjectClause: true}) {
		t.Fatalf("map availability did not match graph where/project scope")
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.Tabularity != TabularityScalar || len(signature.Parameters) != 2 || signature.Parameters[1].Name != "expression" || signature.Parameters[1].TypeKind != ParameterTypeScalar || signature.Parameters[1].ArgumentKind != ArgumentKindExpressionParameter0Element {
		t.Fatalf("map signature = %#v", signature)
	}
	if got := signature.CustomReturnType(CustomReturnTypeContext{ArgumentTypes: []*TypeSymbol{ScalarTypes.Dynamic.Type(), ScalarTypes.String.Type()}}); got != ScalarTypes.DynamicArrayOfString {
		t.Fatalf("map custom return = %T %#v", got, got)
	}
	inner := NewDynamicArraySymbol(ScalarTypes.Long.Type())
	outer, ok := signature.CustomReturnType(CustomReturnTypeContext{ArgumentTypes: []*TypeSymbol{ScalarTypes.Dynamic.Type(), inner.Type()}}).(*DynamicArraySymbol)
	if !ok || outer.ElementType != inner.Type() {
		t.Fatalf("nested map custom return = %T %#v, want exact inner array type %#v", outer, outer, inner.Type())
	}
	if got := signature.CustomReturnType(CustomReturnTypeContext{}); got != ScalarTypes.DynamicArray {
		t.Fatalf("map fallback return = %T %#v", got, got)
	}
}

func checkGraphInnerNodesFunction(t *testing.T) {
	t.Helper()
	fn := Functions.InnerNodes
	if fn.SymbolName() != "inner_nodes" || fn.IsConstantFoldable() || fn.IsHidden() || fn.ResultNameKind() != ResultNameNone || fn.CustomAvailability() == nil {
		t.Fatalf("inner_nodes metadata = name:%s fold:%t hidden:%t result:%s availability:%t", fn.SymbolName(), fn.IsConstantFoldable(), fn.IsHidden(), fn.ResultNameKind(), fn.CustomAvailability() != nil)
	}
	if fn.IsAvailable(CustomAvailabilityContext{}) || fn.IsAvailable(CustomAvailabilityContext{InGraphMatchOrShortestPaths: true}) || fn.IsAvailable(CustomAvailabilityContext{InGraphAllAnyMapFirstArgument: true}) || !fn.IsAvailable(CustomAvailabilityContext{InGraphMatchOrShortestPaths: true, InGraphAllAnyMapFirstArgument: true}) {
		t.Fatalf("inner_nodes availability did not match graph lambda scope")
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeCustom || signature.Tabularity != TabularityScalar || len(signature.Parameters) != 1 || signature.Parameters[0].Name != "edge" || signature.Parameters[0].TypeKind != ParameterTypeScalar {
		t.Fatalf("inner_nodes signature = %#v", signature)
	}
	if got := signature.CustomReturnType(CustomReturnTypeContext{}); got != ScalarTypes.DynamicArrayOfBag {
		t.Fatalf("inner_nodes fallback return = %T %#v", got, got)
	}
}

func checkGeoFunctions(t *testing.T) {
	t.Helper()
	geo := ScalarTypes.GeoShape.Type()
	real := ScalarTypes.Real.Type()
	boolType := ScalarTypes.Bool.Type()
	stringType := ScalarTypes.String.Type()
	intType := ScalarTypes.Int.Type()
	stringArray := ScalarTypes.DynamicArrayOfString.Type()
	stringArrayArray := ScalarTypes.DynamicArrayOfArrayOfString.Type()

	tests := []struct {
		fn         *FunctionSymbol
		name       string
		returnType *TypeSymbol
		minArgs    int
		maxArgs    int
		hidden     bool
	}{
		{Functions.GeoFromWkt, "geo_from_wkt", geo, 1, 1, false},
		{Functions.GeoAngle, "geo_angle", real, 6, 6, false},
		{Functions.GeoAzimuth, "geo_azimuth", real, 4, 4, false},
		{Functions.GeoClosestPointOnLine, "geo_closest_point_on_line", geo, 3, 3, false},
		{Functions.GeoClosestPointOnPolygon, "geo_closest_point_on_polygon", geo, 3, 3, false},
		{Functions.GeoDistance2Points, "geo_distance_2points", real, 4, 5, false},
		{Functions.GeoDistancePointToLine, "geo_distance_point_to_line", real, 3, 3, false},
		{Functions.GeoDistancePointToPolygon, "geo_distance_point_to_polygon", real, 3, 3, false},
		{Functions.GeoPointInCircle, "geo_point_in_circle", boolType, 5, 5, false},
		{Functions.GeoPointInPolygon, "geo_point_in_polygon", boolType, 3, 3, false},
		{Functions.GeoIntersects2Lines, "geo_intersects_2lines", boolType, 2, 2, false},
		{Functions.GeoIntersectsLineWithPolygon, "geo_intersects_line_with_polygon", boolType, 2, 2, false},
		{Functions.GeoIntersects2Polygons, "geo_intersects_2polygons", boolType, 2, 2, false},
		{Functions.GeoIntersection2Lines, "geo_intersection_2lines", geo, 2, 2, false},
		{Functions.GeoIntersectionLineWithPolygon, "geo_intersection_line_with_polygon", geo, 2, 2, false},
		{Functions.GeoIntersection2Polygons, "geo_intersection_2polygons", geo, 2, 2, false},
		{Functions.GeoPolygonsUnion, "geo_union_polygons_array", geo, 1, 1, false},
		{Functions.GeoSimplifyPolygonsArray, "geo_simplify_polygons_array", geo, 1, 2, false},
		{Functions.GeoLinesUnion, "geo_union_lines_array", geo, 1, 1, false},
		{Functions.GeoPolygonToH3Cells, "geo_polygon_to_h3cells", stringArray, 1, 3, false},
		{Functions.GeoPolygonToS2Cells, "geo_polygon_to_s2cells", stringArray, 1, 3, false},
		{Functions.GeoPolygonS2CellCoveringLevel, "__geo_polygon_s2cell_covering_level", intType, 1, 1, true},
		{Functions.GeoPolygonDensify, "geo_polygon_densify", geo, 1, 3, false},
		{Functions.GeoPolygonArea, "geo_polygon_area", real, 1, 1, false},
		{Functions.GeoPolygonBuffer, "geo_polygon_buffer", geo, 2, 3, false},
		{Functions.GeoPolygonCentroid, "geo_polygon_centroid", geo, 1, 1, false},
		{Functions.GeoPolygonValidate, "__geo_polygon_validate", stringType, 1, 1, true},
		{Functions.GeoPolygonPerimeter, "geo_polygon_perimeter", real, 1, 1, false},
		{Functions.GeoPolygonSimplify, "geo_polygon_simplify", geo, 1, 2, false},
		{Functions.GeoLineS2CellCoveringLevel, "__geo_line_s2cell_covering_level", intType, 1, 1, true},
		{Functions.GeoLineLength, "geo_line_length", real, 1, 1, false},
		{Functions.GeoLineBuffer, "geo_line_buffer", geo, 2, 3, false},
		{Functions.GeoLineCentroid, "geo_line_centroid", geo, 1, 1, false},
		{Functions.GeoLineDensify, "geo_line_densify", geo, 1, 3, false},
		{Functions.GeoLineSimplify, "geo_line_simplify", geo, 1, 2, false},
		{Functions.GeoLineLocatePoint, "geo_line_locate_point", real, 3, 4, false},
		{Functions.GeoLineInterpolatePoint, "geo_line_interpolate_point", geo, 2, 2, false},
		{Functions.GeoLineValidate, "__geo_line_validate", stringType, 1, 1, true},
		{Functions.GeoLineToS2Cells, "geo_line_to_s2cells", stringArray, 1, 3, false},
		{Functions.GeoLengthToS2CellLevel, "__geo_length_to_s2cell_level", intType, 1, 1, true},
		{Functions.GeoPointToGeohash, "geo_point_to_geohash", stringType, 2, 3, false},
		{Functions.GeohashToCentralPoint, "geo_geohash_to_central_point", geo, 1, 1, false},
		{Functions.GeohashToPolygon, "geo_geohash_to_polygon", geo, 1, 1, false},
		{Functions.GeohashNeighbors, "geo_geohash_neighbors", stringArray, 1, 1, false},
		{Functions.GeoPointBuffer, "geo_point_buffer", geo, 3, 4, false},
		{Functions.GeoPointToS2Cell, "geo_point_to_s2cell", stringType, 2, 3, false},
		{Functions.GeoS2CellToCentralPoint, "geo_s2cell_to_central_point", geo, 1, 1, false},
		{Functions.GeoS2CellToPolygon, "geo_s2cell_to_polygon", geo, 1, 1, false},
		{Functions.GeoS2CellNeighbors, "geo_s2cell_neighbors", stringArray, 1, 1, false},
		{Functions.GeoPointToH3Cell, "geo_point_to_h3cell", stringType, 2, 3, false},
		{Functions.GeoH3CellToCentralPoint, "geo_h3cell_to_central_point", geo, 1, 1, false},
		{Functions.GeoH3CellToPolygon, "geo_h3cell_to_polygon", geo, 1, 1, false},
		{Functions.GeoH3CellNeighbors, "geo_h3cell_neighbors", stringArray, 1, 1, false},
		{Functions.GeoH3CellChildren, "geo_h3cell_children", stringArray, 1, 2, false},
		{Functions.GeoH3CellParent, "geo_h3cell_parent", stringType, 1, 2, false},
		{Functions.GeoH3CellRings, "geo_h3cell_rings", stringArrayArray, 2, 2, false},
		{Functions.GeoH3CellLevel, "geo_h3cell_level", intType, 1, 1, false},
	}
	for _, tt := range tests {
		signature := tt.fn.Signatures()[0]
		if tt.fn.SymbolName() != tt.name || !tt.fn.IsConstantFoldable() || tt.fn.ResultNameKind() != ResultNameNone || tt.fn.IsHidden() != tt.hidden {
			t.Fatalf("%s metadata = name:%s fold:%t result:%s hidden:%t", tt.name, tt.fn.SymbolName(), tt.fn.IsConstantFoldable(), tt.fn.ResultNameKind(), tt.fn.IsHidden())
		}
		if signature.DeclaredReturnType != tt.returnType || signature.MinArgumentCount != tt.minArgs || signature.MaxArgumentCount != tt.maxArgs {
			t.Fatalf("%s signature = %#v", tt.name, signature)
		}
	}

	checkGeoParameterKind(t, Functions.GeoFromWkt, 0, "wkt", ParameterTypeStringOrArray, false)
	checkGeoParameterKind(t, Functions.GeoDistance2Points, 4, "use_spheroid", ParameterTypeNumberOrBool, true)
	checkGeoParameterKind(t, Functions.GeoPolygonDensify, 2, "preserve_crossing", ParameterTypeNumberOrBool, true)
	checkGeoParameterKind(t, Functions.GeoLineDensify, 2, "preserve_crossing", ParameterTypeNumberOrBool, true)
	checkGeoParameterKind(t, Functions.GeoLineLocatePoint, 3, "use_spheroid", ParameterTypeNumberOrBool, true)
	checkGeoParameterKind(t, Functions.GeoPolygonToH3Cells, 1, "resolution", ParameterTypeNumber, true)
	checkGeoParameterKind(t, Functions.GeoPolygonToH3Cells, 2, "radius", ParameterTypeNumber, true)
	checkGeoParameterKind(t, Functions.GeoPolygonToS2Cells, 1, "level", ParameterTypeNumber, true)
	checkGeoParameterKind(t, Functions.GeoPolygonToS2Cells, 2, "radius", ParameterTypeNumber, true)
	checkGeoParameterKind(t, Functions.GeoLineToS2Cells, 1, "level", ParameterTypeNumber, true)
	checkGeoParameterKind(t, Functions.GeoPointToGeohash, 2, "accuracy", ParameterTypeNumber, true)
	checkGeoParameterKind(t, Functions.GeoPointToS2Cell, 2, "level", ParameterTypeNumber, true)
	checkGeoParameterKind(t, Functions.GeoPointToH3Cell, 2, "resolution", ParameterTypeNumber, true)
	checkGeoParameterDeclared(t, Functions.GeohashToCentralPoint, 0, "geohash", ScalarTypes.String.Type(), false)
	checkGeoParameterDeclared(t, Functions.GeoS2CellToCentralPoint, 0, "s2cell", ScalarTypes.String.Type(), false)
	checkGeoParameterDeclared(t, Functions.GeoH3CellChildren, 0, "h3cell", ScalarTypes.String.Type(), false)
	checkGeoParameterKind(t, Functions.GeoH3CellChildren, 1, "resolution", ParameterTypeNumber, true)
	checkGeoParameterDeclared(t, Functions.GeoH3CellRings, 0, "h3cell", ScalarTypes.String.Type(), false)
	checkGeoParameterKind(t, Functions.GeoH3CellRings, 1, "distance", ParameterTypeNumber, false)
}

func checkGeoParameterKind(t *testing.T, fn *FunctionSymbol, index int, name string, kind ParameterTypeKind, optional bool) {
	t.Helper()
	parameter := fn.Signatures()[0].Parameters[index]
	if parameter.Name != name || parameter.TypeKind != kind || parameter.IsOptional() != optional {
		t.Fatalf("%s parameter %d = %#v", fn.SymbolName(), index, parameter)
	}
}

func checkGeoParameterDeclared(t *testing.T, fn *FunctionSymbol, index int, name string, typ *TypeSymbol, optional bool) {
	t.Helper()
	parameter := fn.Signatures()[0].Parameters[index]
	if parameter.Name != name || parameter.DeclaredTypes[0] != typ || parameter.IsOptional() != optional {
		t.Fatalf("%s parameter %d = %#v", fn.SymbolName(), index, parameter)
	}
}

func checkIPAndFormattingFunctions(t *testing.T) {
	t.Helper()
	checkIPCompareFunction(t, Functions.Ipv4Compare, "ipv4_compare", ScalarTypes.Long.Type())
	checkIPCompareFunction(t, Functions.Ipv4IsMatch, "ipv4_is_match", ScalarTypes.Bool.Type())
	checkIPCompareFunction(t, Functions.Ipv6Compare, "ipv6_compare", ScalarTypes.Long.Type())
	checkIPCompareFunction(t, Functions.Ipv6IsMatch, "ipv6_is_match", ScalarTypes.Bool.Type())
	checkIPStringBoolFunction(t, Functions.Ipv4IsPrivate, "ipv4_is_private", "ip")
	checkIPRangeFunction(t, Functions.Ipv4IsInRange, "ipv4_is_in_range")
	checkIPRangeFunction(t, Functions.Ipv6IsInRange, "ipv6_is_in_range")
	checkIPInAnyRangeFunction(t, Functions.Ipv4IsInAnyRange, "ipv4_is_in_any_range")
	checkIPInAnyRangeFunction(t, Functions.Ipv6IsInAnyRange, "ipv6_is_in_any_range")
	checkIPStringLongFunction(t, Functions.Ipv4NetmaskSuffix, "ipv4_netmask_suffix", "ip")
	checkIpv6LookupRangesFunction(t)
	checkFormatIPV4Function(t)
	checkFormatIPV4MaskFunction(t)
	checkFormatBytesFunction(t)
}

func checkIPCompareFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != returnType || signature.MinArgumentCount != 2 || signature.MaxArgumentCount != 3 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	if signature.Parameters[0].Name != "ip1" || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.String.Type() || signature.Parameters[1].Name != "ip2" || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.String.Type() {
		t.Fatalf("%s ip parameters = %#v", name, signature.Parameters)
	}
	if signature.Parameters[2].Name != "prefix" || signature.Parameters[2].DeclaredTypes[0] != ScalarTypes.Long.Type() || !signature.Parameters[2].IsOptional() {
		t.Fatalf("%s prefix parameter = %#v", name, signature.Parameters[2])
	}
}

func checkIPStringBoolFunction(t *testing.T, fn *FunctionSymbol, name string, parameterName string) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Bool.Type() || signature.Parameters[0].Name != parameterName || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.String.Type() {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkIPStringLongFunction(t *testing.T, fn *FunctionSymbol, name string, parameterName string) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Long.Type() || signature.Parameters[0].Name != parameterName || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.String.Type() {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkIPRangeFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Bool.Type() || signature.Parameters[0].Name != "ip" || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.String.Type() || signature.Parameters[1].Name != "ip_range" || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.String.Type() {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkIPInAnyRangeFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 2 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	if signatures[0].DeclaredReturnType != ScalarTypes.Bool.Type() || signatures[0].Parameters[0].TypeKind != ParameterTypeStringOrDynamic || signatures[0].Parameters[1].DeclaredTypes[0] != ScalarTypes.String.Type() || signatures[0].Parameters[1].MaxOccurring != MaxRepeat {
		t.Fatalf("%s repeated string signature = %#v", name, signatures[0])
	}
	if signatures[1].DeclaredReturnType != ScalarTypes.Bool.Type() || signatures[1].Parameters[0].TypeKind != ParameterTypeStringOrDynamic || signatures[1].Parameters[1].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("%s dynamic array signature = %#v", name, signatures[1])
	}
}

func checkIpv6LookupRangesFunction(t *testing.T) {
	t.Helper()
	fn := Functions.Ipv6LookupRanges
	if fn.SymbolName() != "__ipv6_lookup_ranges" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone || !fn.IsHidden() {
		t.Fatalf("__ipv6_lookup_ranges metadata = name:%s fold:%t result:%s hidden:%t", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.IsHidden())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.DynamicArray.Type() || signature.Parameters[0].TypeKind != ParameterTypeStringOrDynamic || signature.Parameters[1].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("__ipv6_lookup_ranges signature = %#v", signature)
	}
}

func checkFormatIPV4Function(t *testing.T) {
	t.Helper()
	fn := Functions.FormatIPV4
	if fn.SymbolName() != "format_ipv4" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("format_ipv4 metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.String.Type() || signature.MinArgumentCount != 1 || signature.MaxArgumentCount != 2 || signature.Parameters[0].TypeKind != ParameterTypeScalar || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.Long.Type() || !signature.Parameters[1].IsOptional() {
		t.Fatalf("format_ipv4 signature = %#v", signature)
	}
}

func checkFormatIPV4MaskFunction(t *testing.T) {
	t.Helper()
	fn := Functions.FormatIPV4Mask
	if fn.SymbolName() != "format_ipv4_mask" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("format_ipv4_mask metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.String.Type() || signature.MinArgumentCount != 2 || signature.MaxArgumentCount != 2 || signature.Parameters[0].TypeKind != ParameterTypeScalar || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.Long.Type() {
		t.Fatalf("format_ipv4_mask signature = %#v", signature)
	}
}

func checkFormatBytesFunction(t *testing.T) {
	t.Helper()
	fn := Functions.FormatBytes
	if fn.SymbolName() != "format_bytes" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameFirstArgument {
		t.Fatalf("format_bytes metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.String.Type() || signature.MinArgumentCount != 1 || signature.MaxArgumentCount != 3 || signature.Parameters[0].TypeKind != ParameterTypeNumber {
		t.Fatalf("format_bytes signature = %#v", signature)
	}
	if signature.Parameters[1].TypeKind != ParameterTypeNumber || signature.Parameters[1].ArgumentKind != ArgumentKindConstant || !signature.Parameters[1].IsOptional() || signature.Parameters[2].DeclaredTypes[0] != ScalarTypes.String.Type() || signature.Parameters[2].ArgumentKind != ArgumentKindConstant || !signature.Parameters[2].IsOptional() {
		t.Fatalf("format_bytes optional parameters = %#v", signature.Parameters)
	}
}

func checkOtherFunctions(t *testing.T) {
	t.Helper()
	checkSimpleScalarFunction(t, Functions.CurrentClusterEndpoint, "current_cluster_endpoint", ScalarTypes.String.Type(), false, false, ResultNameNone)
	checkSimpleScalarFunction(t, Functions.CurrentDatabase, "current_database", ScalarTypes.String.Type(), false, false, ResultNameNone)
	checkSimpleScalarFunction(t, Functions.CurrentPrincipal, "current_principal", ScalarTypes.String.Type(), false, false, ResultNameNone)
	checkCurrentPrincipalDetailsFunction(t)
	checkCurrentPrincipalIsMemberOfFunction(t)
	checkSimpleScalarFunction(t, Functions.ExtentId, "extent_id", ScalarTypes.Guid.Type(), false, false, ResultNameNone)
	checkObsoleteFunction(t, Functions.ExtentId2, "extentid", "extent_id")
	checkSimpleScalarFunction(t, Functions.ExtentTags, "extent_tags", ScalarTypes.DynamicArrayOfString.Type(), false, false, ResultNameNone)
	checkSimpleScalarFunction(t, Functions.CurrentNodeId, "current_node_id", ScalarTypes.String.Type(), false, true, ResultNameNone)
	checkIngestionTimeFunction(t)
	checkCursorFunction(t, Functions.CursorAfter, "cursor_after", ScalarTypes.Bool.Type(), false)
	checkCursorFunction(t, Functions.CursorBeforeOrAt, "cursor_before_or_at", ScalarTypes.Bool.Type(), false)
	checkSimpleScalarFunction(t, Functions.CursorCurrent, "cursor_current", ScalarTypes.String.Type(), false, false, ResultNameNone)
	checkObsoleteFunction(t, Functions.CursorCurrent2, "current_cursor", "cursor_current")
	checkHasIpv4Function(t, Functions.HasIpv4, "has_ipv4", "ip")
	checkHasIpv4Function(t, Functions.HasIpv4Prefix, "has_ipv4_prefix", "ip_prefix")
	checkHasAnyIpv4Function(t, Functions.HasAnyIpv4, "has_any_ipv4", "ips")
	checkHasAnyIpv4Function(t, Functions.HasAnyIpv4Prefix, "has_any_ipv4_prefix", "ip_prefixes")
	checkSimpleDeclaredFunction(t, Functions.Ipv4RangeToCidrList, "ipv4_range_to_cidr_list", ScalarTypes.DynamicArray.Type(), true, false, ResultNameNone, []string{"start_ip", "end_ip"}, []*TypeSymbol{ScalarTypes.String.Type(), ScalarTypes.String.Type()})
	checkRowNumberFunction(t)
	checkRowCumSumFunction(t)
	checkRowRankFunction(t, Functions.RowRank, "row_rank", "row_rank_dense", "column")
	checkRowRankFunction(t, Functions.RowRankDense, "row_rank_dense", "", "term")
	checkRowRankFunction(t, Functions.RowRankMin, "row_rank_min", "", "term")
	checkRowWindowSessionFunction(t)
	checkRowNavigationFunction(t, Functions.Prev, "prev")
	checkRowNavigationFunction(t, Functions.Next, "next")
	checkSimpleScalarFunction(t, Functions.RowstoreOrdinalRange, "rowstore_ordinal_range", ScalarTypes.Dynamic.Type(), false, true, ResultNameNone)
	checkEstimateDataSizeFunction(t)
	checkSimpleScalarFunction(t, Functions.NewGuid, "new_guid", ScalarTypes.Guid.Type(), false, false, ResultNameNone)
	checkInvokeFunction(t)
	checkCastFunction(t)
	checkSimpleDeclaredFunction(t, Functions.IpGeoLocation, "geo_info_from_ip_address", ScalarTypes.Dynamic.Type(), true, false, ResultNameNone, []string{"ip"}, []*TypeSymbol{ScalarTypes.String.Type()})
	checkColumnNamesOfFunction(t)
}

func checkSimpleScalarFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, foldable bool, hidden bool, resultName ResultNameKind) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() != foldable || fn.IsHidden() != hidden || fn.ResultNameKind() != resultName {
		t.Fatalf("%s metadata = name:%s fold:%t hidden:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.IsHidden(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != returnType || signature.MinArgumentCount != 0 || signature.MaxArgumentCount != 0 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkSimpleDeclaredFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, foldable bool, hidden bool, resultName ResultNameKind, parameterNames []string, parameterTypes []*TypeSymbol) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() != foldable || fn.IsHidden() != hidden || fn.ResultNameKind() != resultName {
		t.Fatalf("%s metadata = name:%s fold:%t hidden:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.IsHidden(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != returnType || signature.MinArgumentCount != len(parameterNames) || signature.MaxArgumentCount != len(parameterNames) {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	for i, parameterName := range parameterNames {
		if signature.Parameters[i].Name != parameterName || signature.Parameters[i].DeclaredTypes[0] != parameterTypes[i] {
			t.Fatalf("%s parameter %d = %#v", name, i, signature.Parameters[i])
		}
	}
}

func checkObsoleteFunction(t *testing.T, fn *FunctionSymbol, name string, alternative string) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsHidden() || fn.Alternative() != alternative || fn.IsConstantFoldable() {
		t.Fatalf("%s obsolete metadata = name:%s hidden:%t alt:%q fold:%t", name, fn.SymbolName(), fn.IsHidden(), fn.Alternative(), fn.IsConstantFoldable())
	}
}

func checkCurrentPrincipalDetailsFunction(t *testing.T) {
	t.Helper()
	fn := Functions.CurrentPrincipalDetails
	if fn.SymbolName() != "current_principal_details" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("current_principal_details metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	bag, ok := signature.DeclaredReturnSymbol.(*DynamicBagSymbol)
	if !ok {
		t.Fatalf("current_principal_details return = %T", signature.DeclaredReturnSymbol)
	}
	assertColumnNames(t, bag.Properties(), []string{"UserPrincipalName", "IdentityProvider", "Authority", "Mfa", "Type", "DisplayName", "ObjectId", "FQN", "Notes"})
}

func checkCurrentPrincipalIsMemberOfFunction(t *testing.T) {
	t.Helper()
	fn := Functions.CurrentPrincipalIsMemberOf
	if fn.SymbolName() != "current_principal_is_member_of" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("current_principal_is_member_of metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	parameter := fn.Signatures()[0].Parameters[0]
	if parameter.Name != "group" || parameter.TypeKind != ParameterTypeStringOrArray || parameter.MinOccurring != 1 || parameter.MaxOccurring != 64 {
		t.Fatalf("current_principal_is_member_of parameter = %#v", parameter)
	}
}

func checkIngestionTimeFunction(t *testing.T) {
	t.Helper()
	fn := Functions.IngestionTime
	if fn.SymbolName() != "ingestion_time" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNamePrefixOnly || fn.ResultNamePrefix() != "$IngestionTime" {
		t.Fatalf("ingestion_time metadata = name:%s fold:%t result:%s prefix:%q", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind(), fn.ResultNamePrefix())
	}
	if fn.Signatures()[0].DeclaredReturnType != ScalarTypes.DateTime.Type() {
		t.Fatalf("ingestion_time signature = %#v", fn.Signatures()[0])
	}
}

func checkCursorFunction(t *testing.T, fn *FunctionSymbol, name string, returnType *TypeSymbol, foldable bool) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() != foldable || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != returnType || signature.Parameters[0].Name != "cursor" || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.String.Type() {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkHasIpv4Function(t *testing.T, fn *FunctionSymbol, name string, secondParameterName string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Bool.Type() || signature.Parameters[0].TypeKind != ParameterTypeStringOrDynamic || signature.Parameters[1].Name != secondParameterName || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.String.Type() {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkHasAnyIpv4Function(t *testing.T, fn *FunctionSymbol, name string, parameterName string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 2 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	if signatures[0].Parameters[0].TypeKind != ParameterTypeStringOrDynamic || signatures[0].Parameters[1].Name != parameterName || signatures[0].Parameters[1].DeclaredTypes[0] != ScalarTypes.String.Type() || signatures[0].Parameters[1].MaxOccurring != MaxRepeat {
		t.Fatalf("%s repeated signature = %#v", name, signatures[0])
	}
	if signatures[1].Parameters[0].TypeKind != ParameterTypeStringOrDynamic || signatures[1].Parameters[1].Name != parameterName || signatures[1].Parameters[1].TypeKind != ParameterTypeDynamicArray {
		t.Fatalf("%s dynamic signature = %#v", name, signatures[1])
	}
}

func checkRowNumberFunction(t *testing.T) {
	t.Helper()
	fn := Functions.RowNumber
	if fn.SymbolName() != "row_number" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("row_number metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Long.Type() || signature.MinArgumentCount != 0 || signature.MaxArgumentCount != 2 || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.Long.Type() || !signature.Parameters[0].IsOptional() || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.Bool.Type() || !signature.Parameters[1].IsOptional() {
		t.Fatalf("row_number signature = %#v", signature)
	}
}

func checkRowCumSumFunction(t *testing.T) {
	t.Helper()
	fn := Functions.RowCumSum
	if fn.SymbolName() != "row_cumsum" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("row_cumsum metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeParameter0 || signature.Parameters[0].Name != "term" || signature.Parameters[0].TypeKind != ParameterTypeNumber || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.Bool.Type() || !signature.Parameters[1].IsOptional() {
		t.Fatalf("row_cumsum signature = %#v", signature)
	}
}

func checkRowRankFunction(t *testing.T, fn *FunctionSymbol, name string, alternative string, firstParameterName string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	if alternative != "" {
		if !fn.IsHidden() || fn.Alternative() != alternative {
			t.Fatalf("%s obsolete metadata = hidden:%t alt:%q", name, fn.IsHidden(), fn.Alternative())
		}
	} else if fn.IsHidden() {
		t.Fatalf("%s should not be hidden", name)
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.Long.Type() || signature.Parameters[0].Name != firstParameterName || signature.Parameters[0].TypeKind != ParameterTypeNotDynamic || signature.Parameters[1].IsOptional() == false {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkRowWindowSessionFunction(t *testing.T) {
	t.Helper()
	fn := Functions.RowWindowSession
	if fn.SymbolName() != "row_window_session" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("row_window_session metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeParameter0 || signature.MinArgumentCount != 3 || signature.MaxArgumentCount != 4 || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.DateTime.Type() || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.TimeSpan.Type() || signature.Parameters[2].DeclaredTypes[0] != ScalarTypes.TimeSpan.Type() || signature.Parameters[3].DeclaredTypes[0] != ScalarTypes.Bool.Type() || !signature.Parameters[3].IsOptional() {
		t.Fatalf("row_window_session signature = %#v", signature)
	}
}

func checkRowNavigationFunction(t *testing.T, fn *FunctionSymbol, name string) {
	t.Helper()
	if fn.SymbolName() != name || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeParameter0 || signature.Parameters[0].ArgumentKind != ArgumentKindColumn || signature.Parameters[0].TypeKind != ParameterTypeScalar || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.Long.Type() || !signature.Parameters[1].IsOptional() || signature.Parameters[2].TypeKind != ParameterTypeScalar || signature.Parameters[2].ArgumentKind != ArgumentKindConstant || !signature.Parameters[2].IsOptional() {
		t.Fatalf("%s signature = %#v", name, signature)
	}
}

func checkEstimateDataSizeFunction(t *testing.T) {
	t.Helper()
	fn := Functions.EstimateDataSize
	if fn.SymbolName() != "estimate_data_size" || fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("estimate_data_size metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 2 {
		t.Fatalf("estimate_data_size signatures = %d", len(signatures))
	}
	if signatures[0].DeclaredReturnType != ScalarTypes.Long.Type() || signatures[0].Parameters[0].TypeKind != ParameterTypeScalar || signatures[0].Parameters[0].ArgumentKind != ArgumentKindColumn || signatures[0].Parameters[0].MaxOccurring != MaxRepeat {
		t.Fatalf("estimate_data_size column signature = %#v", signatures[0])
	}
	if signatures[1].DeclaredReturnType != ScalarTypes.Long.Type() || signatures[1].Parameters[0].ArgumentKind != ArgumentKindStarOnly {
		t.Fatalf("estimate_data_size star signature = %#v", signatures[1])
	}
}

func checkInvokeFunction(t *testing.T) {
	t.Helper()
	fn := Functions.Invoke
	if fn.SymbolName() != "__invoke" || !fn.IsConstantFoldable() || !fn.IsHidden() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("__invoke metadata = name:%s fold:%t hidden:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.IsHidden(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeParameter1Literal || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.String.Type() || signature.Parameters[0].ArgumentKind != ArgumentKindLiteral || signature.Parameters[1].DeclaredTypes[0] != ScalarTypes.TypeLit.Type() || signature.Parameters[1].ArgumentKind != ArgumentKindLiteral || signature.Parameters[2].TypeKind != ParameterTypeAny || signature.Parameters[2].MinOccurring != 0 || signature.Parameters[2].MaxOccurring != MaxRepeat {
		t.Fatalf("__invoke signature = %#v", signature)
	}
}

func checkCastFunction(t *testing.T) {
	t.Helper()
	fn := Functions.Cast
	if fn.SymbolName() != "__cast" || !fn.IsConstantFoldable() || !fn.IsHidden() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("__cast metadata = name:%s fold:%t hidden:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.IsHidden(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.ReturnKind != ReturnTypeParameter0Literal || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.TypeLit.Type() || signature.Parameters[0].ArgumentKind != ArgumentKindLiteral || signature.Parameters[1].TypeKind != ParameterTypeScalar {
		t.Fatalf("__cast signature = %#v", signature)
	}
}

func checkColumnNamesOfFunction(t *testing.T) {
	t.Helper()
	fn := Functions.ColumnNamesOf
	if fn.SymbolName() != "column_names_of" || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("column_names_of metadata = name:%s fold:%t result:%s", fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signature := fn.Signatures()[0]
	if signature.DeclaredReturnType != ScalarTypes.DynamicArrayOfString.Type() || signature.Parameters[0].Name != "table" || signature.Parameters[0].TypeKind != ParameterTypeTabular {
		t.Fatalf("column_names_of signature = %#v", signature)
	}
}

func assertColumnNames(t *testing.T, columns []*ColumnSymbol, want []string) {
	t.Helper()
	if len(columns) != len(want) {
		t.Fatalf("column count = %d, want %d: %#v", len(columns), len(want), columns)
	}
	for i, name := range want {
		if columns[i] == nil || columns[i].SymbolName() != name {
			t.Fatalf("column %d = %#v, want %q", i, columns[i], name)
		}
	}
}

func checkConvertFunction(t *testing.T, fn *FunctionSymbol, name string, values int, first string, last string) {
	t.Helper()
	if fn.SymbolName() != name || !fn.IsConstantFoldable() || fn.ResultNameKind() != ResultNameNone {
		t.Fatalf("%s metadata = name:%s fold:%t result:%s", name, fn.SymbolName(), fn.IsConstantFoldable(), fn.ResultNameKind())
	}
	signatures := fn.Signatures()
	if len(signatures) != 1 {
		t.Fatalf("%s signatures = %d", name, len(signatures))
	}
	signature := signatures[0]
	if signature.DeclaredReturnType != ScalarTypes.Real.Type() || signature.MinArgumentCount != 3 || signature.MaxArgumentCount != 3 {
		t.Fatalf("%s signature = %#v", name, signature)
	}
	if len(signature.Parameters) != 3 {
		t.Fatalf("%s parameters = %#v", name, signature.Parameters)
	}
	if signature.Parameters[0].Name != "value" || signature.Parameters[0].DeclaredTypes[0] != ScalarTypes.Real.Type() {
		t.Fatalf("%s value parameter = %#v", name, signature.Parameters[0])
	}
	for _, parameter := range signature.Parameters[1:] {
		if parameter.ArgumentKind != ArgumentKindExpression || len(parameter.Values) != values {
			t.Fatalf("%s %s parameter values = %#v", name, parameter.Name, parameter.Values)
		}
		if parameter.Values[0] != first || parameter.Values[len(parameter.Values)-1] != last {
			t.Fatalf("%s %s value bounds = %v %v", name, parameter.Name, parameter.Values[0], parameter.Values[len(parameter.Values)-1])
		}
	}
}

func TestConvertFunctionsAreAvailableThroughGlobalState(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db"))
	if len(globals.Functions) != len(Functions.All) {
		t.Fatalf("global function count = %d", len(globals.Functions))
	}
	if globals.GetFunction("tostring") != Functions.ToString {
		t.Fatal("tostring was not found through GlobalState")
	}
	if globals.GetFunction("convert_length") != Functions.ConvertLength {
		t.Fatal("convert_length was not found through GlobalState")
	}
	if globals.GetFunction("base64_encode_tostring") != Functions.Base64EncodeToString {
		t.Fatal("base64_encode_tostring was not found through GlobalState")
	}
	if globals.GetFunction("parse_url") != Functions.ParseUrl {
		t.Fatal("parse_url was not found through GlobalState")
	}
	if globals.GetFunction("datetime_diff") != Functions.DatetimeDiff {
		t.Fatal("datetime_diff was not found through GlobalState")
	}
	if globals.GetFunction("datetime_list_timezones") != Functions.DateTimeListTimezones {
		t.Fatal("datetime_list_timezones was not found through GlobalState")
	}
	if globals.GetFunction("hash_xxhash64") != Functions.HashXXH64 {
		t.Fatal("hash_xxhash64 was not found through GlobalState")
	}
	if globals.GetFunction("__hash_crc32") != Functions.HashCrc32 {
		t.Fatal("__hash_crc32 was not found through GlobalState")
	}
	if globals.GetFunction("case") != Functions.Case {
		t.Fatal("case was not found through GlobalState")
	}
	if globals.GetFunction("bin_at") != Functions.BinAt {
		t.Fatal("bin_at was not found through GlobalState")
	}
	if globals.GetFunction("column_ifexists") != Functions.ColumnIfExists {
		t.Fatal("column_ifexists was not found through GlobalState")
	}
	if globals.GetFunction("binary_shift_left") != Functions.BinaryShiftLeft {
		t.Fatal("binary_shift_left was not found through GlobalState")
	}
	if globals.GetFunction("array_length") != Functions.ArrayLength {
		t.Fatal("array_length was not found through GlobalState")
	}
	if globals.GetFunction("array_sort_asc") != Functions.ArraySortAsc {
		t.Fatal("array_sort_asc was not found through GlobalState")
	}

	local := NewScalarFunctionSymbol("convert_length", ScalarTypes.String.Type())
	globals = NewGlobalState(NewDatabaseSymbol("db", local))
	if globals.GetAnyFunction("convert_length") != local {
		t.Fatal("database function should take precedence over built-in conversion function")
	}

	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db")))
	info := binder.ResolveFunction("convert_temperature")
	if info.ReferencedSignature() == nil || info.ReferencedSignature().DeclaredReturnType != ScalarTypes.Real.Type() {
		t.Fatalf("convert_temperature binding info = %#v", info)
	}
	if diags := binder.ValidateFunctionArgumentCount("convert_temperature", 2); len(diags) != 1 || diags[0].Code != "KS119" {
		t.Fatalf("convert_temperature argument diagnostics = %#v", diags)
	}
}
