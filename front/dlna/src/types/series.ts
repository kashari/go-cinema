export interface Episode {
  ID: string;
  Title: string;
  Description: string;
  Path: string;
  ResumeAt: string;
  EpisodeIndex: number;
  SeriesID: string;
}

export interface Series {
  ID: string;
  Title: string;
  Description: string;
  BaseDir: string;
  Episodes: Episode[];
  CurrentIndex: number;
}
